package main

import "time"

// Notifier interface to notify when a product becomes available or is sold out again
type Notifier interface {
	NotifyWhenAvailable(string, string, float64, string, string) error
	NotifyWhenNotAvailable(string, time.Duration) error
}
