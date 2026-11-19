package database

import "time"

const (
	RoleAdmin int16 = 1
	RoleUser  int16 = 2

	BookingConfirmed int16 = 1
	BookingCancelled int16 = 2
)

type User struct {
	ID           int64     `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Phone        *string   `json:"phone"`
	Role         int16     `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Event struct {
	ID             int64     `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Location       string    `json:"location"`
	StartsAt       time.Time `json:"starts_at"`
	TotalSeats     int       `json:"total_seats"`
	AvailableSeats int       `json:"available_seats"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Booking struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	EventID     int64      `json:"event_id"`
	Quantity    int16      `json:"quantity"`
	Status      int16      `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CancelledAt *time.Time `json:"cancelled_at"`
}

func ValidRole(role int16) bool { return role == RoleAdmin || role == RoleUser }

func ValidBookingStatus(status int16) bool {
	return status == BookingConfirmed || status == BookingCancelled
}
