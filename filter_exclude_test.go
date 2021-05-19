package main

import (
	"fmt"
	"testing"
)

func TestExcludeFilter(t *testing.T) {
	tests := []struct {
		regex    string // exclusive regex
		name     string // product name
		included bool   // should be included or not
	}{
		{"(?i)(rtx|rx)(.*)(3060|3070|3080|3090|5700|6800|6900)( )?(xt|ti)?", "MSI GeForce RTX 3060 GAMING X", false},           // 3060 in the exclude regex
		{"(?i)(rtx|rx)(.*)(3060|3070|3080|3090|5700|6800|6900)( )?(xt|ti)?", "ASUS AMD Radeon RX 5600 XT TUF Gaming X3", true}, // 5600 not in the exclude regex
		{"", "MSI GeForce RTX 3060 GAMING X", true}, // do nothing when the exclude regex is empty
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("TestExcludeFilter#%d", i), func(t *testing.T) {
			product := &Product{Name: tc.name}
			filter, err := NewExcludeFilter(tc.regex)
			if err != nil {
				t.Errorf("cannot create filter with regex '%s': %s", tc.regex, err)
			}

			included := filter.Include(product)

			if included != tc.included {
				t.Errorf("regex '%s' for product '%s': got included=%t, want included=%t", tc.regex, tc.name, included, tc.included)
			} else {
				if included {
					t.Logf("regex '%s' includes product '%s'", tc.regex, tc.name)
				} else {
					t.Logf("regex '%s' excludes product '%s'", tc.regex, tc.name)
				}
			}

		})
	}
}
