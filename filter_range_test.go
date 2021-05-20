package main

import (
	"fmt"
	"testing"
)

func TestRangeFilter(t *testing.T) {
	tests := []struct {
		name     string  // product name
		price    float64 // product price
		model    string  // model regex to apply on the product name
		min      float64 // minimum price
		max      float64 // maximum price
		included bool    // should be included or not
	}{
		{"MSI GeForce RTX 3090 GAMING X", 99.99, "3090", 50.0, 100.0, true},   // model match and price is in the range, should be included
		{"MSI GeForce RTX 3090 GAMING X", 99.99, "3080", 50.0, 100.0, true},   // model doesn't match, should be included
		{"MSI GeForce RTX 3090 GAMING X", 999.99, "3090", 50.0, 100.0, false}, // model match and price is outside of the range, shoud not be included
		{"MSI GeForce RTX 3090 GAMING X", 99.99, "", 50.0, 100.0, true},       // model regex is missing, should be included
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("TestRangeFilter#%d", i), func(t *testing.T) {
			product := &Product{Name: tc.name, Price: tc.price}
			filter, err := NewRangeFilter(tc.model, tc.min, tc.max)
			if err != nil {
				t.Errorf("cannot create filter with model regex '%s' and price range [%.2f, %.2f]: %s", tc.model, tc.min, tc.max, err)
			}

			included := filter.Include(product)

			if included != tc.included {
				t.Errorf("product '%s' with model regex '%s' and range [%.2f, %.2f]: got included=%t, want included=%t", tc.name, tc.model, tc.min, tc.max, included, tc.included)
			} else {
				if included {
					t.Logf("product '%s' included by model regex '%s' and range [%.2f, %.2f]", tc.name, tc.model, tc.min, tc.max)
				} else {
					t.Logf("product '%s' excluded by model regex '%s' and range [%.2f, %.2f]", tc.name, tc.model, tc.min, tc.max)
				}
			}

		})
	}
}
