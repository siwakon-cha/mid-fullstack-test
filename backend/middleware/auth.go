package middleware

import (
	"net/http"
	"strings"

	"github.com/Terminal-cmd99/mid-fullstack-test/backend/database"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const CookieName = "event_session"

type Claims struct {
	Role int16 `json:"role"`
	jwt.RegisteredClaims
}

func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

func RequireAuth(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, err := c.Cookie(CookieName)
		if err != nil {
			Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "กรุณาเข้าสู่ระบบ")
			return
		}
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(value, claims, func(token *jwt.Token) (any, error) {
			return secret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || !token.Valid || !database.ValidRole(claims.Role) {
			Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "เซสชันไม่ถูกต้องหรือหมดอายุ")
			return
		}
		id, err := claims.GetSubject()
		if err != nil || strings.TrimSpace(id) == "" {
			Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "เซสชันไม่ถูกต้อง")
			return
		}
		c.Set("user_id", id)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RequireAdmin(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || role.(int16) != database.RoleAdmin {
		Error(c, http.StatusForbidden, "FORBIDDEN", "ต้องใช้สิทธิ์ผู้ดูแลระบบ")
		return
	}
	c.Next()
}

func RequireOrigin(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		origin := strings.TrimSuffix(c.GetHeader("Origin"), "/")
		if origin == "" || origin != strings.TrimSuffix(allowedOrigin, "/") {
			Error(c, http.StatusForbidden, "CSRF_ORIGIN", "origin ของคำขอไม่ถูกต้อง")
			return
		}
		c.Next()
	}
}
