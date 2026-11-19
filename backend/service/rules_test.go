package service

import "testing"

func TestValidateQuantity(t *testing.T) {
	for _, quantity := range []int{-1, 0, 5, 100} {
		if ValidateQuantity(quantity) == nil {
			t.Fatalf("quantity %d should be rejected", quantity)
		}
	}
	for _, quantity := range []int{1, 4} {
		if err := ValidateQuantity(quantity); err != nil {
			t.Fatalf("quantity %d should be valid: %v", quantity, err)
		}
	}
}

func TestAvailableAfterCapacityChange(t *testing.T) {
	available, err := AvailableAfterCapacityChange(100, 70, 150)
	if err != nil || available != 120 {
		t.Fatalf("expected 120 available, got %d, err=%v", available, err)
	}
	if _, err := AvailableAfterCapacityChange(100, 70, 20); err == nil {
		t.Fatal("capacity below confirmed seats should be rejected")
	}
}

func TestNormalizeAndEscapeSearch(t *testing.T) {
	if got := NormalizeSearch("  MUSIC  "); got != "music" {
		t.Fatalf("unexpected normalized search: %q", got)
	}
	if got := EscapeLike(`50%_off`); got != `50\%\_off` {
		t.Fatalf("wildcards were not escaped: %q", got)
	}
}
