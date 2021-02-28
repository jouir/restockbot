package main

import (
	"fmt"
	"testing"
)

func TestBuildHashtags(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"MSI GeForce RTX 3060 GAMING X", "#nvidia #rtx3060"},
		{"MSI GeForce RTX 3060 Ti GAMING X", "#nvidia #rtx3060ti"}, // with space (3060 Ti)
		{"MSI RTX 3060Ti VENTUS 2X OC", "#nvidia #rtx3060ti"},      // without space (3060Ti)
		{"MSI GeForce RTX 3070 GAMING TRIO", "#nvidia #rtx3070"},
		{"MSI GeForce RTX 3080 SUPRIM X", "#nvidia #rtx3080"},
		{"MSI GeForce RTX 3090 GAMING X TRIO 24G", "#nvidia #rtx3090"},
		{"MSI Radeon RX 5700 XT GAMING X", "#amd #rx5700xt"},                      // with space (5700 XT)
		{"ASUS Radeon RX 5700XT ROG-STRIX-RX5700XT-O8G-GAMING", "#amd #rx5700xt"}, // without space (5700XT)
		{"MSI Radeon RX 6800", "#amd #rx6800"},
		{"MSI Radeon RX 6800 XT", "#amd #rx6800xt"},                   // with space (6800 XT)
		{"POWERCOLOR RX 6800XT Red Dragon", "#amd #rx6800xt"},         // without space (6800XT)
		{"MSI Radeon RX 6900 XT GAMING X TRIO 16G", "#amd #rx6900xt"}, // with space (6900 XT)
		{"POWERCOLOR RED DEVIL RX 6900XT 16GB", "#amd #rx6900xt"},     // without space (6900XT)
		{"unknown product", ""},
		{"", ""},
	}

	notifier := TwitterNotifier{
		client: nil,
		user:   nil,
		db:     nil,
		hashtagsMap: []map[string]string{
			{"rtx 3060( )?ti": "#nvidia #rtx3060ti"},
			{"rtx 3060": "#nvidia #rtx3060"},
			{"rtx 3070": "#nvidia #rtx3070"},
			{"rtx 3080": "#nvidia #rtx3080"},
			{"rtx 3090": "#nvidia #rtx3090"},
			{"rx 6800( )?xt": "#amd #rx6800xt"},
			{"rx 6900( )?xt": "#amd #rx6900xt"},
			{"rx 6800": "#amd #rx6800"},
			{"rx 5700( )?xt": "#amd #rx5700xt"},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("TestBuildHashtags#%d", i), func(t *testing.T) {
			got := notifier.buildHashtags(tc.input)
			if got != tc.expected {
				t.Errorf("for %s, got %s, want %s", tc.input, got, tc.expected)
			} else {
				t.Logf("for %s, want %s, got %s", tc.input, got, tc.expected)
			}
		})
	}
}

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
