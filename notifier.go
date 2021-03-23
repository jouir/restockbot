package main

import (
	"fmt"
	"time"
)

// Notifier interface to notify when a product becomes available or is sold out again
type Notifier interface {
	NotifyWhenAvailable(string, string, float64, string, string) error
	NotifyWhenNotAvailable(string, time.Duration) error
}

// formatPrice using internationalization rules
// euro sign is placed after the value
// default the currency, or symbol if applicable, is placed before the value
func formatPrice(value float64, currency string) string {
	switch {
	case currency == "EUR":
		return fmt.Sprintf("%.2f€", value)
	case currency == "USD":
		return fmt.Sprintf("$%.2f", value)
	default:
		return fmt.Sprintf("%s%.2f", currency, value)
	}
}
