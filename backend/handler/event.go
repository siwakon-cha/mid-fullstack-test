package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	appcache "github.com/Terminal-cmd99/mid-fullstack-test/backend/cache"
	"github.com/Terminal-cmd99/mid-fullstack-test/backend/database"
	"github.com/Terminal-cmd99/mid-fullstack-test/backend/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Event struct {
	DB    *gorm.DB
	Cache *appcache.Events
}

type eventInput struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartsAt    time.Time `json:"starts_at"`
	TotalSeats  int       `json:"total_seats"`
}

func (h Event) List(c *gin.Context) {
	page, err1 := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, err2 := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err1 != nil || err2 != nil || page < 1 || pageSize < 1 || pageSize > 50 {
		respondError(c, http.StatusBadRequest, "INVALID_PAGINATION", "page ต้องมากกว่า 0 และ pageSize ต้องอยู่ระหว่าง 1–50")
		return
	}
	search := service.NormalizeSearch(c.Query("search"))
	key := h.Cache.Key(c, page, pageSize, search)
	if value, ok := h.Cache.Get(c, key); ok {
		c.Header("X-Cache", "HIT")
		c.Data(http.StatusOK, "application/json; charset=utf-8", value)
		return
	}

	query := h.DB.WithContext(c).Model(&database.Event{}).Where("starts_at > NOW()")
	if search != "" {
		query = query.Where(`LOWER(title) LIKE ? ESCAPE '\'`, service.EscapeLike(search)+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL", "ไม่สามารถอ่านรายการอีเวนต์ได้")
		return
	}
	var events []database.Event
	if err := query.Order("starts_at, id").Offset((page - 1) * pageSize).Limit(pageSize).Find(&events).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL", "ไม่สามารถอ่านรายการอีเวนต์ได้")
		return
	}
	response := gin.H{"items": events, "page": page, "page_size": pageSize, "total": total}
	encoded, _ := json.Marshal(response)
	h.Cache.Set(c, key, encoded)
	c.Header("X-Cache", "MISS")
	c.Data(http.StatusOK, "application/json; charset=utf-8", encoded)
}

func (h Event) Detail(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var event database.Event
	if err := h.DB.WithContext(c).First(&event, id).Error; err != nil {
		respondError(c, http.StatusNotFound, "EVENT_NOT_FOUND", "ไม่พบอีเวนต์")
		return
	}
	c.JSON(http.StatusOK, gin.H{"event": event})
}

func validateEventInput(input eventInput) error {
	if strings.TrimSpace(input.Title) == "" || len(strings.TrimSpace(input.Title)) > 200 {
		return service.ErrCapacity
	}
	if input.TotalSeats <= 0 || !input.StartsAt.After(time.Now()) {
		return service.ErrCapacity
	}
	return nil
}

func (h Event) Create(c *gin.Context) {
	var input eventInput
	if err := c.ShouldBindJSON(&input); err != nil || validateEventInput(input) != nil {
		respondError(c, http.StatusBadRequest, "INVALID_EVENT", "ชื่อ เวลาในอนาคต และจำนวนที่นั่งต้องถูกต้อง")
		return
	}
	event := database.Event{Title: strings.TrimSpace(input.Title), Description: strings.TrimSpace(input.Description), Location: strings.TrimSpace(input.Location), StartsAt: input.StartsAt, TotalSeats: input.TotalSeats, AvailableSeats: input.TotalSeats}
	if err := h.DB.WithContext(c).Create(&event).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL", "ไม่สามารถสร้างอีเวนต์ได้")
		return
	}
	h.Cache.Invalidate(c)
	c.JSON(http.StatusCreated, gin.H{"event": event})
}

func (h Event) Update(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var input eventInput
	if err := c.ShouldBindJSON(&input); err != nil || validateEventInput(input) != nil {
		respondError(c, http.StatusBadRequest, "INVALID_EVENT", "ชื่อ เวลาในอนาคต และจำนวนที่นั่งต้องถูกต้อง")
		return
	}
	err := h.DB.WithContext(c).Transaction(func(tx *gorm.DB) error {
		var updatedID int64
		if err := tx.Raw(`UPDATE events SET available_seats = ? - (total_seats - available_seats), total_seats = ?, updated_at = NOW() WHERE id = ? AND ? >= (total_seats - available_seats) RETURNING id`, input.TotalSeats, input.TotalSeats, id, input.TotalSeats).Row().Scan(&updatedID); err != nil {
			return service.ErrCapacity
		}
		return tx.Model(&database.Event{}).Where("id = ?", id).Updates(map[string]any{
			"title": strings.TrimSpace(input.Title), "description": strings.TrimSpace(input.Description), "location": strings.TrimSpace(input.Location), "starts_at": input.StartsAt,
		}).Error
	})
	if err != nil {
		if err == service.ErrCapacity {
			respondError(c, http.StatusConflict, "CAPACITY_CONFLICT", "จำนวนที่นั่งรวมห้ามต่ำกว่าที่จองแล้ว")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL", "ไม่สามารถแก้ไขอีเวนต์ได้")
		return
	}
	h.Cache.Invalidate(c)
	var event database.Event
	h.DB.First(&event, id)
	c.JSON(http.StatusOK, gin.H{"event": event})
}
