package main

import (
	"fmt"
	"testing"
)

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		value    float64
		currency string
		expected string
	}{
		{999.99, "EUR", "999.99€"},
		{999.99, "USD", "$999.99"},
		{999.99, "CHF", "CHF999.99"},
		{999.99, "", "999.99"},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("TestFormatPrice#%d", i), func(t *testing.T) {
			got := formatPrice(tc.value, tc.currency)
			if got != tc.expected {
				t.Errorf("for value %0.2f and currency %s, got %s, want %s", tc.value, tc.currency, got, tc.expected)
			} else {
				t.Logf("for value %0.2f and currency %s, got %s, want %s", tc.value, tc.currency, got, tc.expected)
			}
		})
	}
}
