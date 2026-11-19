package database

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type DemoCredentials struct {
	AdminEmail    string
	AdminPassword string
	UserEmail     string
	UserPassword  string
}

func SeedDemo(db *gorm.DB, credentials DemoCredentials) error {
	users := []struct {
		FirstName string
		LastName  string
		Email     string
		Password  string
		Role      int16
	}{
		{"Demo", "Admin", credentials.AdminEmail, credentials.AdminPassword, RoleAdmin},
		{"Demo", "User", credentials.UserEmail, credentials.UserPassword, RoleUser},
	}
	for _, item := range users {
		var count int64
		if err := db.Model(&User{}).Where("email = ?", item.Email).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user := User{FirstName: item.FirstName, LastName: item.LastName, Email: item.Email, PasswordHash: string(hash), Role: item.Role}
		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}

	locations := []string{"Bangkok Art Centre", "Queen Sirikit Convention Center", "River City Bangkok", "True Digital Park"}
	for i := 1; i <= 20; i++ {
		title := fmt.Sprintf("Bangkok Event %02d", i)
		var count int64
		if err := db.Model(&Event{}).Where("title = ?", title).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		seats := 40 + (i * 5)
		event := Event{
			Title:          title,
			Description:    "Local demo event for the take-home assignment.",
			Location:       locations[(i-1)%len(locations)],
			StartsAt:       time.Now().UTC().AddDate(0, 1, i),
			TotalSeats:     seats,
			AvailableSeats: seats,
		}
		if err := db.Create(&event).Error; err != nil {
			return err
		}
	}
	return nil
}
