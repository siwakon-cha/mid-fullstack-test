package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func respondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

func pathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "รหัสต้องเป็นจำนวนเต็มบวก")
		return 0, false
	}
	return id, true
}

func currentUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "กรุณาเข้าสู่ระบบ")
		return 0, false
	}
	id, err := strconv.ParseInt(value.(string), 10, 64)
	if err != nil || id <= 0 {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "เซสชันไม่ถูกต้อง")
		return 0, false
	}
	return id, true
}
