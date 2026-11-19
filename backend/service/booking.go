package service

import (
	"context"
	"errors"
	"time"

	appcache "github.com/Terminal-cmd99/mid-fullstack-test/backend/cache"
	"github.com/Terminal-cmd99/mid-fullstack-test/backend/database"
	"gorm.io/gorm"
)

var (
	ErrSoldOut         = errors.New("event not available or seats are insufficient")
	ErrBookingLimit    = errors.New("booking would exceed four seats for this event")
	ErrBookingNotFound = errors.New("booking not found")
	ErrForbidden       = errors.New("booking belongs to another user")
	ErrAlreadyCanceled = errors.New("booking already cancelled")
	ErrCapacity        = errors.New("total seats cannot be lower than confirmed seats")
)

type Booking struct {
	db    *gorm.DB
	cache *appcache.Events
}

func NewBooking(db *gorm.DB, cache *appcache.Events) *Booking {
	return &Booking{db: db, cache: cache}
}

func (s *Booking) Create(ctx context.Context, userID, eventID int64, quantity int) (int64, int, error) {
	if err := ValidateQuantity(quantity); err != nil {
		return 0, 0, err
	}
	var bookingID int64
	var totalQuantity int
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var updatedEventID int64
		row := tx.Raw(`
			UPDATE events
			SET available_seats = available_seats - ?, updated_at = NOW()
			WHERE id = ? AND starts_at > NOW() AND available_seats >= ?
			RETURNING id`, quantity, eventID, quantity).Row()
		if err := row.Scan(&updatedEventID); err != nil {
			return ErrSoldOut
		}

		row = tx.Raw(`
			INSERT INTO bookings (user_id, event_id, quantity, status)
			VALUES (?, ?, ?, ?)
			ON CONFLICT (user_id, event_id)
			DO UPDATE SET quantity = CASE
			                  WHEN bookings.status = ? THEN bookings.quantity + EXCLUDED.quantity
			                  ELSE EXCLUDED.quantity
			              END,
			              status = ?, cancelled_at = NULL, updated_at = NOW()
			WHERE bookings.status = ? OR bookings.quantity + EXCLUDED.quantity <= 4
			RETURNING id, quantity`, userID, eventID, quantity, database.BookingConfirmed,
			database.BookingConfirmed, database.BookingConfirmed, database.BookingCancelled).Row()
		if err := row.Scan(&bookingID, &totalQuantity); err != nil {
			return ErrBookingLimit
		}
		return nil
	})
	if err == nil {
		s.cache.Invalidate(ctx)
	}
	return bookingID, totalQuantity, err
}

func (s *Booking) Cancel(ctx context.Context, userID, bookingID int64) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var eventID int64
		if err := tx.Raw(`SELECT event_id FROM bookings WHERE id = ?`, bookingID).Row().Scan(&eventID); err != nil {
			return ErrBookingNotFound
		}
		var lockedEventID int64
		if err := tx.Raw(`SELECT id FROM events WHERE id = ? FOR UPDATE`, eventID).Row().Scan(&lockedEventID); err != nil {
			return err
		}
		var booking database.Booking
		if err := tx.Raw(`SELECT * FROM bookings WHERE id = ? FOR UPDATE`, bookingID).Scan(&booking).Error; err != nil || booking.ID == 0 {
			return ErrBookingNotFound
		}
		if booking.UserID != userID {
			return ErrForbidden
		}
		if booking.Status == database.BookingCancelled {
			return ErrAlreadyCanceled
		}
		now := time.Now().UTC()
		result := tx.Model(&database.Booking{}).Where("id = ? AND status = ?", bookingID, database.BookingConfirmed).
			Updates(map[string]any{"status": database.BookingCancelled, "cancelled_at": now, "updated_at": now})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrAlreadyCanceled
		}
		return tx.Model(&database.Event{}).Where("id = ?", eventID).
			Updates(map[string]any{"available_seats": gorm.Expr("available_seats + ?", booking.Quantity), "updated_at": now}).Error
	})
	if err == nil {
		s.cache.Invalidate(ctx)
	}
	return err
}

func (s *Booking) UpdateCapacity(ctx context.Context, eventID int64, newTotal int) error {
	var id int64
	row := s.db.WithContext(ctx).Raw(`
		UPDATE events
		SET available_seats = ? - (total_seats - available_seats),
		    total_seats = ?, updated_at = NOW()
		WHERE id = ? AND ? >= (total_seats - available_seats) AND ? > 0
		RETURNING id`, newTotal, newTotal, eventID, newTotal, newTotal).Row()
	if err := row.Scan(&id); err != nil {
		return ErrCapacity
	}
	s.cache.Invalidate(ctx)
	return nil
}
