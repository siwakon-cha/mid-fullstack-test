package database

import "testing"

func TestRoleAndBookingStatusConstants(t *testing.T) {
	if RoleAdmin != 1 || RoleUser != 2 || !ValidRole(RoleAdmin) || !ValidRole(RoleUser) || ValidRole(3) {
		t.Fatal("role constants or validation are incorrect")
	}
	if BookingConfirmed != 1 || BookingCancelled != 2 || !ValidBookingStatus(BookingConfirmed) || !ValidBookingStatus(BookingCancelled) || ValidBookingStatus(0) {
		t.Fatal("booking status constants or validation are incorrect")
	}
}
