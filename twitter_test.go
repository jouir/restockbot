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
		{"MSI GeForce RTX 3060 Ti GAMING X", "#nvidia #rtx3060ti"},
		{"MSI GeForce RTX 3070 GAMING TRIO", "#nvidia #rtx3070"},
		{"MSI GeForce RTX 3080 SUPRIM X", "#nvidia #rtx3080"},
		{"MSI GeForce RTX 3090 GAMING X TRIO 24G", "#nvidia #rtx3090"},
		{"MSI Radeon RX 5700 XT GAMING X", "#amd #rx5700xt"},
		{"MSI Radeon RX 6800", "#amd #rx6800"},
		{"MSI Radeon RX 6800 XT", "#amd #rx6800xt"},
		{"MSI Radeon RX 6900 XT GAMING X TRIO 16G", "#amd #rx6900xt"},
		{"unknown product", ""},
		{"", ""},
	}

	notifier := TwitterNotifier{
		client: nil,
		user:   nil,
		db:     nil,
		hashtagsMap: map[string]string{
			"rtx 3060 ti": "#nvidia #rtx3060ti",
			"rtx 3060ti":  "#nvidia #rtx3060ti",
			"rtx 3060":    "#nvidia #rtx3060",
			"rtx 3070":    "#nvidia #rtx3070",
			"rtx 3080":    "#nvidia #rtx3080",
			"rtx 3090":    "#nvidia #rtx3090",
			"rx 6800 xt":  "#amd #rx6800xt",
			"rx 6800xt":   "#amd #rx6800xt",
			"rx 6900xt":   "#amd #rx6900xt",
			"rx 6900 xt":  "#amd #rx6900xt",
			"rx 6800":     "#amd #rx6800",
			"rx 5700 xt":  "#amd #rx5700xt",
			"rx 5700xt":   "#amd #rx5700xt",
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("TestBuildHashtags#%d", i), func(t *testing.T) {
			got := notifier.buildHashtags(tc.input)
			if got != tc.expected {
				t.Errorf("got %s, want %s", got, tc.expected)
			} else {
				t.Logf("success")
			}
		})
	}
}
