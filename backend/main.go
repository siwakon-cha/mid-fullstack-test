package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	appcache "github.com/Terminal-cmd99/mid-fullstack-test/backend/cache"
	"github.com/Terminal-cmd99/mid-fullstack-test/backend/database"
	"github.com/Terminal-cmd99/mid-fullstack-test/backend/handler"
	"github.com/Terminal-cmd99/mid-fullstack-test/backend/middleware"
	"github.com/Terminal-cmd99/mid-fullstack-test/backend/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Config struct {
	Port              string
	DatabaseURL       string
	RedisURL          string
	JWTSecret         string
	FrontendOrigin    string
	AppEnv            string
	SeedDemo          bool
	DemoAdminEmail    string
	DemoAdminPassword string
	DemoUserEmail     string
	DemoUserPassword  string
}

type Application struct {
	DB     *gorm.DB
	Cache  *appcache.Events
	Config Config
}

func loadConfig() Config {
	return Config{
		Port: env("PORT", "8080"), DatabaseURL: os.Getenv("DATABASE_URL"), RedisURL: os.Getenv("REDIS_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"), FrontendOrigin: env("FRONTEND_ORIGIN", "http://localhost:3000"), AppEnv: env("APP_ENV", "development"),
		SeedDemo:       strings.ToLower(env("SEED_DEMO", "true")) == "true",
		DemoAdminEmail: env("DEMO_ADMIN_EMAIL", "admin.demo@example.com"), DemoAdminPassword: env("DEMO_ADMIN_PASSWORD", "LocalAdmin123!"),
		DemoUserEmail: env("DEMO_USER_EMAIL", "user.demo@example.com"), DemoUserPassword: env("DEMO_USER_PASSWORD", "LocalUser123!"),
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func (a Application) Router() *gin.Engine {
	if a.Config.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	if a.Config.AppEnv != "test" {
		router.Use(gin.Logger())
	}
	router.Use(gin.Recovery())

	authHandler := handler.Auth{DB: a.DB, JWTSecret: []byte(a.Config.JWTSecret), CookieSecure: a.Config.AppEnv == "production"}
	eventHandler := handler.Event{DB: a.DB, Cache: a.Cache}
	bookingService := service.NewBooking(a.DB, a.Cache)
	bookingHandler := handler.Booking{DB: a.DB, Service: bookingService}
	authRequired := middleware.RequireAuth([]byte(a.Config.JWTSecret))

	router.GET("/health", func(c *gin.Context) {
		sqlDB, err := a.DB.DB()
		if err != nil || sqlDB.PingContext(c) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	api.Use(middleware.RequireOrigin(a.Config.FrontendOrigin))
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)
	api.GET("/events", eventHandler.List)
	api.GET("/events/:id", eventHandler.Detail)

	protected := api.Group("")
	protected.Use(authRequired)
	protected.POST("/auth/logout", authHandler.Logout)
	protected.GET("/auth/me", authHandler.Me)
	protected.POST("/events/:id/bookings", bookingHandler.Create)
	protected.GET("/bookings/me", bookingHandler.Mine)
	protected.DELETE("/bookings/:id", bookingHandler.Cancel)

	admin := api.Group("/admin")
	admin.Use(authRequired, middleware.RequireAdmin)
	admin.POST("/events", eventHandler.Create)
	admin.PATCH("/events/:id", eventHandler.Update)

	return router
}

func main() {
	config := loadConfig()
	if len(config.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters")
	}
	db, err := connectWithRetry(config.DatabaseURL, 30*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}
	if config.SeedDemo {
		credentials := database.DemoCredentials{AdminEmail: config.DemoAdminEmail, AdminPassword: config.DemoAdminPassword, UserEmail: config.DemoUserEmail, UserPassword: config.DemoUserPassword}
		if err := database.SeedDemo(db, credentials); err != nil {
			log.Fatal(err)
		}
	}
	eventCache, err := appcache.New(config.RedisURL)
	if err != nil {
		log.Printf("Redis configuration ignored: %v", err)
		eventCache, _ = appcache.New("")
	}
	app := Application{DB: db, Cache: eventCache, Config: config}
	server := &http.Server{Addr: ":" + config.Port, Handler: app.Router(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("backend listening on :%s", config.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func connectWithRetry(databaseURL string, timeout time.Duration) (*gorm.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var lastErr error
	for {
		db, err := database.Connect(databaseURL)
		if err == nil {
			sqlDB, sqlErr := db.DB()
			if sqlErr == nil && sqlDB.PingContext(ctx) == nil {
				return db, nil
			}
			lastErr = sqlErr
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("connect PostgreSQL: %w", lastErr)
		case <-time.After(time.Second):
		}
	}
}
