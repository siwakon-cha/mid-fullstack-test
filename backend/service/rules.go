package service

import (
	"fmt"
	"strings"
)

func ValidateQuantity(quantity int) error {
	if quantity < 1 || quantity > 4 {
		return fmt.Errorf("quantity must be between 1 and 4")
	}
	return nil
}

func AvailableAfterCapacityChange(total, available, newTotal int) (int, error) {
	confirmed := total - available
	if newTotal <= 0 || newTotal < confirmed {
		return 0, fmt.Errorf("new total cannot be lower than confirmed seats")
	}
	return newTotal - confirmed, nil
}

func NormalizeSearch(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) > 80 {
		value = value[:80]
	}
	return value
}

func EscapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	return strings.ReplaceAll(value, `_`, `\_`)
}
