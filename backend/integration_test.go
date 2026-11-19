//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	appcache "github.com/Terminal-cmd99/mid-fullstack-test/backend/cache"
	"github.com/Terminal-cmd99/mid-fullstack-test/backend/database"
	"github.com/Terminal-cmd99/mid-fullstack-test/backend/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	testOrigin = "http://integration.test"
	testSecret = "integration-test-secret-at-least-32-characters"
)

type testEnvironment struct {
	db     *gorm.DB
	cache  *appcache.Events
	server *httptest.Server
}

func setupIntegration(t *testing.T) testEnvironment {
	t.Helper()
	gin.SetMode(gin.TestMode)
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" || (!strings.Contains(databaseURL, "_test") && !strings.Contains(databaseURL, "postgres-test")) {
		t.Fatal("TEST_DATABASE_URL must point to an isolated test database containing _test or postgres-test")
	}
	db, err := database.Connect(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("TRUNCATE bookings, events, users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatal(err)
	}
	eventCache, err := appcache.New(os.Getenv("TEST_REDIS_URL"))
	if err != nil {
		t.Fatal(err)
	}
	eventCache.Invalidate(context.Background())
	app := Application{DB: db, Cache: eventCache, Config: Config{JWTSecret: testSecret, FrontendOrigin: testOrigin, AppEnv: "test"}}
	server := httptest.NewServer(app.Router())
	t.Cleanup(server.Close)
	return testEnvironment{db: db, cache: eventCache, server: server}
}

func request(t *testing.T, client *http.Client, method, url string, body any, cookie *http.Cookie, origin string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func closeResponse(t *testing.T, response *http.Response) {
	t.Helper()
	_, _ = io.Copy(io.Discard, response.Body)
	_ = response.Body.Close()
}

func createUser(t *testing.T, db *gorm.DB, index int, role int16) database.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("TestPassword123!"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	user := database.User{FirstName: "Test", LastName: fmt.Sprintf("User%d", index), Email: fmt.Sprintf("user%03d@integration.test", index), PasswordHash: string(hash), Role: role}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	return user
}

func createEvent(t *testing.T, db *gorm.DB, title string, seats int) database.Event {
	t.Helper()
	event := database.Event{Title: title, Description: "integration test", Location: "Test Hall", StartsAt: time.Now().UTC().Add(24 * time.Hour), TotalSeats: seats, AvailableSeats: seats}
	if err := db.Create(&event).Error; err != nil {
		t.Fatal(err)
	}
	return event
}

func authCookie(t *testing.T, user database.User) *http.Cookie {
	t.Helper()
	claims := middleware.Claims{Role: user.Role, RegisteredClaims: jwt.RegisteredClaims{Subject: strconv.FormatInt(user.ID, 10), ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	return &http.Cookie{Name: middleware.CookieName, Value: signed, Path: "/", HttpOnly: true}
}

func TestIntegrationAPI_MainFlows(t *testing.T) {
	env := setupIntegration(t)
	client := &http.Client{Timeout: 10 * time.Second}

	register := request(t, client, http.MethodPost, env.server.URL+"/api/auth/register", map[string]any{
		"first_name": "Normal", "last_name": "User", "email": "normal@integration.test", "password": "TestPassword123!", "role": 1,
	}, nil, testOrigin)
	if register.StatusCode != http.StatusCreated {
		t.Fatalf("register status=%d", register.StatusCode)
	}
	closeResponse(t, register)
	var registered database.User
	if err := env.db.Where("email = ?", "normal@integration.test").First(&registered).Error; err != nil || registered.Role != database.RoleUser || registered.PasswordHash == "TestPassword123!" {
		t.Fatalf("registration did not force role=2 and bcrypt: user=%+v err=%v", registered, err)
	}

	login := request(t, client, http.MethodPost, env.server.URL+"/api/auth/login", map[string]any{"email": registered.Email, "password": "TestPassword123!"}, nil, testOrigin)
	if login.StatusCode != http.StatusOK || len(login.Cookies()) == 0 || !login.Cookies()[0].HttpOnly || login.Cookies()[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("login cookie invalid: status=%d cookies=%v", login.StatusCode, login.Cookies())
	}
	userCookie := login.Cookies()[0]
	closeResponse(t, login)
	me := request(t, client, http.MethodGet, env.server.URL+"/api/auth/me", nil, userCookie, "")
	if me.StatusCode != http.StatusOK {
		t.Fatalf("me status=%d", me.StatusCode)
	}
	closeResponse(t, me)

	csrf := request(t, client, http.MethodPost, env.server.URL+"/api/auth/logout", nil, userCookie, "http://evil.test")
	if csrf.StatusCode != http.StatusForbidden {
		t.Fatalf("invalid origin should be 403, got %d", csrf.StatusCode)
	}
	closeResponse(t, csrf)

	event := createEvent(t, env.db, "Music Integration Night", 8)
	list := request(t, client, http.MethodGet, env.server.URL+"/api/events?page=1&pageSize=10&search=music", nil, nil, "")
	if list.StatusCode != http.StatusOK || list.Header.Get("X-Cache") == "" {
		t.Fatalf("event list failed: status=%d cache=%q", list.StatusCode, list.Header.Get("X-Cache"))
	}
	if list.Header.Get("X-Cache") != "MISS" {
		t.Fatalf("first event list should miss cache, got %q", list.Header.Get("X-Cache"))
	}
	closeResponse(t, list)
	listHit := request(t, client, http.MethodGet, env.server.URL+"/api/events?page=1&pageSize=10&search=music", nil, nil, "")
	if listHit.StatusCode != http.StatusOK || listHit.Header.Get("X-Cache") != "HIT" {
		t.Fatalf("second event list should hit cache: status=%d cache=%q", listHit.StatusCode, listHit.Header.Get("X-Cache"))
	}
	closeResponse(t, listHit)
	detail := request(t, client, http.MethodGet, fmt.Sprintf("%s/api/events/%d", env.server.URL, event.ID), nil, nil, "")
	if detail.StatusCode != http.StatusOK {
		t.Fatalf("event detail status=%d", detail.StatusCode)
	}
	closeResponse(t, detail)

	invalid := request(t, client, http.MethodPost, fmt.Sprintf("%s/api/events/%d/bookings", env.server.URL, event.ID), map[string]int{"quantity": 5}, userCookie, testOrigin)
	if invalid.StatusCode != http.StatusBadRequest {
		t.Fatalf("quantity 5 status=%d", invalid.StatusCode)
	}
	closeResponse(t, invalid)
	book := request(t, client, http.MethodPost, fmt.Sprintf("%s/api/events/%d/bookings", env.server.URL, event.ID), map[string]int{"quantity": 2}, userCookie, testOrigin)
	if book.StatusCode != http.StatusCreated {
		t.Fatalf("book status=%d", book.StatusCode)
	}
	closeResponse(t, book)
	listAfterBook := request(t, client, http.MethodGet, env.server.URL+"/api/events?page=1&pageSize=10&search=music", nil, nil, "")
	if listAfterBook.StatusCode != http.StatusOK || listAfterBook.Header.Get("X-Cache") != "MISS" {
		t.Fatalf("booking should invalidate event cache: status=%d cache=%q", listAfterBook.StatusCode, listAfterBook.Header.Get("X-Cache"))
	}
	closeResponse(t, listAfterBook)
	addSeats := request(t, client, http.MethodPost, fmt.Sprintf("%s/api/events/%d/bookings", env.server.URL, event.ID), map[string]int{"quantity": 2}, userCookie, testOrigin)
	if addSeats.StatusCode != http.StatusCreated {
		t.Fatalf("add seats status=%d", addSeats.StatusCode)
	}
	closeResponse(t, addSeats)
	var booking database.Booking
	env.db.Where("user_id = ? AND event_id = ?", registered.ID, event.ID).First(&booking)
	if booking.Quantity != 4 {
		t.Fatalf("repeated booking should accumulate to 4, got %d", booking.Quantity)
	}
	overLimit := request(t, client, http.MethodPost, fmt.Sprintf("%s/api/events/%d/bookings", env.server.URL, event.ID), map[string]int{"quantity": 1}, userCookie, testOrigin)
	if overLimit.StatusCode != http.StatusConflict {
		t.Fatalf("booking over cumulative limit status=%d", overLimit.StatusCode)
	}
	closeResponse(t, overLimit)
	var afterLimit database.Event
	env.db.First(&afterLimit, event.ID)
	env.db.First(&booking, booking.ID)
	if afterLimit.AvailableSeats != 4 || booking.Quantity != 4 {
		t.Fatalf("booking limit rollback failed: available=%d quantity=%d", afterLimit.AvailableSeats, booking.Quantity)
	}
	cancel := request(t, client, http.MethodDelete, fmt.Sprintf("%s/api/bookings/%d", env.server.URL, booking.ID), nil, userCookie, testOrigin)
	if cancel.StatusCode != http.StatusOK {
		t.Fatalf("cancel status=%d", cancel.StatusCode)
	}
	closeResponse(t, cancel)
	cancelAgain := request(t, client, http.MethodDelete, fmt.Sprintf("%s/api/bookings/%d", env.server.URL, booking.ID), nil, userCookie, testOrigin)
	if cancelAgain.StatusCode != http.StatusConflict {
		t.Fatalf("repeat cancel status=%d", cancelAgain.StatusCode)
	}
	closeResponse(t, cancelAgain)
	rebook := request(t, client, http.MethodPost, fmt.Sprintf("%s/api/events/%d/bookings", env.server.URL, event.ID), map[string]int{"quantity": 1}, userCookie, testOrigin)
	if rebook.StatusCode != http.StatusCreated {
		t.Fatalf("rebook status=%d", rebook.StatusCode)
	}
	closeResponse(t, rebook)

	admin := createUser(t, env.db, 900, database.RoleAdmin)
	userForbidden := request(t, client, http.MethodPost, env.server.URL+"/api/admin/events", map[string]any{"title": "Forbidden", "starts_at": time.Now().Add(48 * time.Hour), "total_seats": 10}, userCookie, testOrigin)
	if userForbidden.StatusCode != http.StatusForbidden {
		t.Fatalf("normal user admin status=%d", userForbidden.StatusCode)
	}
	closeResponse(t, userForbidden)
	adminCreate := request(t, client, http.MethodPost, env.server.URL+"/api/admin/events", map[string]any{"title": "Admin Created", "description": "created", "location": "Hall", "starts_at": time.Now().Add(48 * time.Hour), "total_seats": 10}, authCookie(t, admin), testOrigin)
	if adminCreate.StatusCode != http.StatusCreated {
		t.Fatalf("admin create status=%d", adminCreate.StatusCode)
	}
	closeResponse(t, adminCreate)
	var createdEvent database.Event
	if err := env.db.Where("title = ?", "Admin Created").First(&createdEvent).Error; err != nil {
		t.Fatal(err)
	}
	adminUpdate := request(t, client, http.MethodPatch, fmt.Sprintf("%s/api/admin/events/%d", env.server.URL, createdEvent.ID), map[string]any{"title": "Admin Updated", "description": "updated", "location": "New Hall", "starts_at": time.Now().Add(72 * time.Hour), "total_seats": 15}, authCookie(t, admin), testOrigin)
	if adminUpdate.StatusCode != http.StatusOK {
		t.Fatalf("admin update status=%d", adminUpdate.StatusCode)
	}
	closeResponse(t, adminUpdate)

	invalidRole := database.User{FirstName: "Bad", LastName: "Role", Email: "bad-role@integration.test", PasswordHash: "hash", Role: 3}
	if err := env.db.Create(&invalidRole).Error; err == nil {
		t.Fatal("database role constraint did not reject role=3")
	}

	logout := request(t, client, http.MethodPost, env.server.URL+"/api/auth/logout", nil, userCookie, testOrigin)
	if logout.StatusCode != http.StatusOK || len(logout.Cookies()) == 0 || logout.Cookies()[0].MaxAge >= 0 {
		t.Fatalf("logout did not clear cookie: status=%d cookies=%v", logout.StatusCode, logout.Cookies())
	}
	closeResponse(t, logout)
}

func TestConcurrentBookings_100Users_10Seats(t *testing.T) {
	env := setupIntegration(t)
	event := createEvent(t, env.db, "Concurrency Event", 10)
	cookies := make([]*http.Cookie, 100)
	for i := range cookies {
		cookies[i] = authCookie(t, createUser(t, env.db, i, database.RoleUser))
	}
	var created, conflicts, unexpected atomic.Int64
	start := make(chan struct{})
	var wait sync.WaitGroup
	for i := range cookies {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			<-start
			client := &http.Client{Timeout: 20 * time.Second}
			response := request(t, client, http.MethodPost, fmt.Sprintf("%s/api/events/%d/bookings", env.server.URL, event.ID), map[string]int{"quantity": 1}, cookies[index], testOrigin)
			switch response.StatusCode {
			case http.StatusCreated:
				created.Add(1)
			case http.StatusConflict:
				conflicts.Add(1)
			default:
				unexpected.Add(1)
			}
			closeResponse(t, response)
		}(i)
	}
	close(start)
	wait.Wait()
	var actual database.Event
	env.db.First(&actual, event.ID)
	var confirmed int64
	env.db.Model(&database.Booking{}).Where("event_id = ? AND status = ?", event.ID, database.BookingConfirmed).Select("COALESCE(SUM(quantity), 0)").Scan(&confirmed)
	t.Logf("concurrency result: created=%d conflict=%d unexpected=%d available=%d confirmed=%d total=%d", created.Load(), conflicts.Load(), unexpected.Load(), actual.AvailableSeats, confirmed, actual.TotalSeats)
	if created.Load() != 10 || conflicts.Load() != 90 || unexpected.Load() != 0 || actual.AvailableSeats != 0 || confirmed != 10 || int64(actual.AvailableSeats)+confirmed != int64(actual.TotalSeats) {
		t.Fatal("100-request concurrency invariant failed")
	}
}

func TestConcurrentBookings_SameUserCannotExceedFourSeats(t *testing.T) {
	env := setupIntegration(t)
	event := createEvent(t, env.db, "Same User Limit Event", 10)
	user := createUser(t, env.db, 1, database.RoleUser)
	cookie := authCookie(t, user)

	var created, conflicts, unexpected atomic.Int64
	start := make(chan struct{})
	var wait sync.WaitGroup
	for range 10 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			response := request(t, &http.Client{Timeout: 10 * time.Second}, http.MethodPost, fmt.Sprintf("%s/api/events/%d/bookings", env.server.URL, event.ID), map[string]int{"quantity": 1}, cookie, testOrigin)
			switch response.StatusCode {
			case http.StatusCreated:
				created.Add(1)
			case http.StatusConflict:
				conflicts.Add(1)
			default:
				unexpected.Add(1)
			}
			closeResponse(t, response)
		}()
	}
	close(start)
	wait.Wait()

	var actualEvent database.Event
	var booking database.Booking
	env.db.First(&actualEvent, event.ID)
	env.db.Where("user_id = ? AND event_id = ?", user.ID, event.ID).First(&booking)
	t.Logf("same-user limit result: created=%d conflict=%d unexpected=%d available=%d quantity=%d", created.Load(), conflicts.Load(), unexpected.Load(), actualEvent.AvailableSeats, booking.Quantity)
	if created.Load() != 4 || conflicts.Load() != 6 || unexpected.Load() != 0 || actualEvent.AvailableSeats != 6 || booking.Quantity != 4 {
		t.Fatal("concurrent repeated bookings exceeded the four-seat user/event limit")
	}
}

func TestConcurrentCancellationReturnsSeatsOnce(t *testing.T) {
	env := setupIntegration(t)
	client := &http.Client{Timeout: 10 * time.Second}
	user := createUser(t, env.db, 1, database.RoleUser)
	event := createEvent(t, env.db, "Cancel Race Event", 10)
	cookie := authCookie(t, user)
	book := request(t, client, http.MethodPost, fmt.Sprintf("%s/api/events/%d/bookings", env.server.URL, event.ID), map[string]int{"quantity": 3}, cookie, testOrigin)
	if book.StatusCode != http.StatusCreated {
		t.Fatalf("book before cancel race status=%d", book.StatusCode)
	}
	closeResponse(t, book)
	var booking database.Booking
	env.db.Where("user_id = ? AND event_id = ?", user.ID, event.ID).First(&booking)

	var okCount, conflictCount atomic.Int64
	start := make(chan struct{})
	var wait sync.WaitGroup
	for range 10 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			response := request(t, &http.Client{Timeout: 10 * time.Second}, http.MethodDelete, fmt.Sprintf("%s/api/bookings/%d", env.server.URL, booking.ID), nil, cookie, testOrigin)
			if response.StatusCode == http.StatusOK {
				okCount.Add(1)
			} else if response.StatusCode == http.StatusConflict {
				conflictCount.Add(1)
			}
			closeResponse(t, response)
		}()
	}
	close(start)
	wait.Wait()
	var actual database.Event
	env.db.First(&actual, event.ID)
	t.Logf("cancel race result: ok=%d conflict=%d available=%d", okCount.Load(), conflictCount.Load(), actual.AvailableSeats)
	if okCount.Load() != 1 || conflictCount.Load() != 9 || actual.AvailableSeats != 10 {
		t.Fatal("concurrent cancellation returned seats more than once")
	}
}

func TestRedisUnavailableFallsBackToPostgres(t *testing.T) {
	env := setupIntegration(t)
	event := createEvent(t, env.db, "Redis Fallback Event", 5)
	user := createUser(t, env.db, 1, database.RoleUser)
	brokenCache, err := appcache.New("redis://127.0.0.1:1/0")
	if err != nil {
		t.Fatal(err)
	}
	app := Application{DB: env.db, Cache: brokenCache, Config: Config{JWTSecret: testSecret, FrontendOrigin: testOrigin, AppEnv: "test"}}
	server := httptest.NewServer(app.Router())
	defer server.Close()
	response := request(t, &http.Client{Timeout: 5 * time.Second}, http.MethodGet, server.URL+"/api/events", nil, nil, "")
	defer closeResponse(t, response)
	if response.StatusCode != http.StatusOK || response.Header.Get("X-Cache") != "MISS" {
		t.Fatalf("Redis fallback failed: status=%d cache=%q", response.StatusCode, response.Header.Get("X-Cache"))
	}
	book := request(t, &http.Client{Timeout: 5 * time.Second}, http.MethodPost, fmt.Sprintf("%s/api/events/%d/bookings", server.URL, event.ID), map[string]int{"quantity": 1}, authCookie(t, user), testOrigin)
	defer closeResponse(t, book)
	if book.StatusCode != http.StatusCreated {
		t.Fatalf("booking should succeed while Redis is offline, status=%d", book.StatusCode)
	}
}
