package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/Terminal-cmd99/mid-fullstack-test/backend/database"
	"github.com/Terminal-cmd99/mid-fullstack-test/backend/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Booking struct {
	DB      *gorm.DB
	Service *service.Booking
}

func (h Booking) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	eventID, ok := pathID(c)
	if !ok {
		return
	}
	var input struct {
		Quantity int `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || service.ValidateQuantity(input.Quantity) != nil {
		respondError(c, http.StatusBadRequest, "INVALID_QUANTITY", "จำนวนต้องอยู่ระหว่าง 1–4")
		return
	}
	id, totalQuantity, err := h.Service.Create(c, userID, eventID, input.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingLimit):
			respondError(c, http.StatusConflict, "BOOKING_LIMIT", "ยอดจองรวมต่ออีเวนต์ต้องไม่เกิน 4 ที่นั่ง")
		case errors.Is(err, service.ErrSoldOut):
			respondError(c, http.StatusConflict, "SOLD_OUT", "อีเวนต์หมดเวลา ที่นั่งไม่พอ หรือไม่พบอีเวนต์")
		default:
			respondError(c, http.StatusInternalServerError, "INTERNAL", "ไม่สามารถจองได้")
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"booking": gin.H{"id": id, "event_id": eventID, "quantity": totalQuantity, "status": database.BookingConfirmed}})
}

func (h Booking) Mine(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	rows := make([]struct {
		database.Booking
		EventTitle    string    `json:"event_title"`
		EventStartsAt time.Time `json:"event_starts_at"`
		EventLocation string    `json:"event_location"`
	}, 0)
	err := h.DB.WithContext(c).Table("bookings").Select("bookings.*, events.title AS event_title, events.starts_at AS event_starts_at, events.location AS event_location").
		Joins("JOIN events ON events.id = bookings.event_id").Where("bookings.user_id = ?", userID).
		Order("bookings.created_at DESC, bookings.id DESC").Scan(&rows).Error
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL", "ไม่สามารถอ่านรายการจองได้")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

func (h Booking) Cancel(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	err := h.Service.Cancel(c, userID, id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			respondError(c, http.StatusForbidden, "FORBIDDEN", "ไม่สามารถยกเลิกการจองของผู้อื่นได้")
		case errors.Is(err, service.ErrAlreadyCanceled):
			respondError(c, http.StatusConflict, "ALREADY_CANCELLED", "การจองถูกยกเลิกแล้ว")
		case errors.Is(err, service.ErrBookingNotFound):
			respondError(c, http.StatusNotFound, "BOOKING_NOT_FOUND", "ไม่พบการจอง")
		default:
			respondError(c, http.StatusInternalServerError, "INTERNAL", "ไม่สามารถยกเลิกการจองได้")
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ยกเลิกการจองแล้ว"})
}
