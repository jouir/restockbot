package main

import (
	"fmt"
	"regexp"
	"testing"
)

func TestFilterInclusive(t *testing.T) {
	tests := []struct {
		regex    string // inclusive regex
		name     string // product name
		included bool   // should be included or not
	}{
		{"(?i)(rtx|rx)(.*)(3060|3070|3080|3090|5700|6800|6900)( )?(xt|ti)?", "MSI GeForce RTX 3060 GAMING X", true},             // 3060 in the include regex
		{"(?i)(rtx|rx)(.*)(3060|3070|3080|3090|5700|6800|6900)( )?(xt|ti)?", "ASUS AMD Radeon RX 5600 XT TUF Gaming X3", false}, // 5600 not in the include regex
		{"", "MSI GeForce RTX 3060 GAMING X", true}, // do nothing when the include regex is empty
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("TestFilterInclusive#%d", i), func(t *testing.T) {
			var regexCompiled *regexp.Regexp
			var err error
			if tc.regex != "" {
				regexCompiled, err = regexp.Compile(tc.regex)
				if err != nil {
					t.Errorf("cannot parse regex '%s'", tc.regex)
				}
			}
			products := []*Product{{Name: tc.name}}
			filtered := filterInclusive(regexCompiled, products)
			included := false
			for _, product := range filtered {
				if product.Name == tc.name && !included {
					included = true
				}
			}
			if tc.included != included {
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

func TestFilterExclusive(t *testing.T) {
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
		t.Run(fmt.Sprintf("TestFilterExclusive#%d", i), func(t *testing.T) {
			var regexCompiled *regexp.Regexp
			var err error
			if tc.regex != "" {
				regexCompiled, err = regexp.Compile(tc.regex)
				if err != nil {
					t.Errorf("cannot parse regex '%s'", tc.regex)
				}
			}

			products := []*Product{{Name: tc.name}}
			filtered := filterExclusive(regexCompiled, products)
			included := false
			for _, product := range filtered {
				if product.Name == tc.name && !included {
					included = true
				}
			}
			if tc.included != included {
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
