package handler

import (
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/Terminal-cmd99/mid-fullstack-test/backend/database"
	"github.com/Terminal-cmd99/mid-fullstack-test/backend/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Auth struct {
	DB           *gorm.DB
	JWTSecret    []byte
	CookieSecure bool
}

type registerInput struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	Phone     *string `json:"phone"`
}

func (h Auth) Register(c *gin.Context) {
	var input registerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_JSON", "ข้อมูล JSON ไม่ถูกต้อง")
		return
	}
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.FirstName == "" || len(input.FirstName) > 100 || input.LastName == "" || len(input.LastName) > 100 {
		respondError(c, http.StatusBadRequest, "INVALID_NAME", "กรุณากรอกชื่อและนามสกุล")
		return
	}
	if address, err := mail.ParseAddress(input.Email); err != nil || address.Address != input.Email {
		respondError(c, http.StatusBadRequest, "INVALID_EMAIL", "อีเมลไม่ถูกต้อง")
		return
	}
	if len(input.Password) < 8 || len(input.Password) > 72 {
		respondError(c, http.StatusBadRequest, "INVALID_PASSWORD", "รหัสผ่านต้องยาว 8–72 ตัวอักษร")
		return
	}
	if input.Phone != nil {
		trimmed := strings.TrimSpace(*input.Phone)
		if len(trimmed) > 20 {
			respondError(c, http.StatusBadRequest, "INVALID_PHONE", "เบอร์โทรยาวเกินไป")
			return
		}
		if trimmed == "" {
			input.Phone = nil
		} else {
			input.Phone = &trimmed
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL", "ไม่สามารถสร้างบัญชีได้")
		return
	}
	user := database.User{FirstName: input.FirstName, LastName: input.LastName, Email: input.Email, PasswordHash: string(hash), Phone: input.Phone, Role: database.RoleUser}
	if err := h.DB.Create(&user).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			respondError(c, http.StatusConflict, "EMAIL_EXISTS", "อีเมลนี้ถูกใช้งานแล้ว")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL", "ไม่สามารถสร้างบัญชีได้")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (h Auth) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || input.Email == "" || input.Password == "" {
		respondError(c, http.StatusBadRequest, "INVALID_CREDENTIALS", "กรุณากรอกอีเมลและรหัสผ่าน")
		return
	}
	var user database.User
	if err := h.DB.Where("email = ?", strings.ToLower(strings.TrimSpace(input.Email))).First(&user).Error; err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		respondError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
		return
	}
	now := time.Now()
	claims := middleware.Claims{Role: user.Role, RegisteredClaims: jwt.RegisteredClaims{
		Subject: strconv.FormatInt(user.ID, 10), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(8 * time.Hour)),
	}}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(h.JWTSecret)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL", "ไม่สามารถเข้าสู่ระบบได้")
		return
	}
	h.setCookie(c, signed, int((8 * time.Hour).Seconds()))
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h Auth) Logout(c *gin.Context) {
	h.setCookie(c, "", -1)
	c.JSON(http.StatusOK, gin.H{"message": "ออกจากระบบแล้ว"})
}

func (h Auth) Me(c *gin.Context) {
	id, ok := currentUserID(c)
	if !ok {
		return
	}
	var user database.User
	if err := h.DB.First(&user, id).Error; err != nil {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "ไม่พบบัญชีผู้ใช้")
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h Auth) setCookie(c *gin.Context, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.CookieName, value, maxAge, "/", "", h.CookieSecure, true)
}
