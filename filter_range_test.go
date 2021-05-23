package main

import (
	"fmt"
	"testing"
)

func TestRangeFilter(t *testing.T) {
	tests := []struct {
		product  *Product
		model    string  // model regex to apply on the product name
		min      float64 // minimum price
		max      float64 // maximum price
		currency string  // price currency
		included bool    // should be included or not
	}{
		{&Product{Name: "MSI GeForce RTX 3090 GAMING X", Price: 99.99, PriceCurrency: "EUR"}, "3090", 50.0, 100.0, "EUR", true},   // model match and price is in the range, should be included
		{&Product{Name: "MSI GeForce RTX 3090 GAMING X", Price: 99.99, PriceCurrency: "EUR"}, "3080", 50.0, 100.0, "EUR", true},   // model doesn't match, should be included
		{&Product{Name: "MSI GeForce RTX 3090 GAMING X", Price: 999.99, PriceCurrency: "EUR"}, "3090", 50.0, 100.0, "EUR", false}, // model match and price is outside of the range, shoud not be included
		{&Product{Name: "MSI GeForce RTX 3090 GAMING X", Price: 99.99, PriceCurrency: "EUR"}, "", 50.0, 100.0, "EUR", true},       // model regex is missing, should be included
		{&Product{Name: "MSI GeForce RTX 3090 GAMING X", Price: 99.99, PriceCurrency: "EUR"}, "3090", 50.0, 0.0, "EUR", true},     // upper limit is missing, should be included
	}

	converter := NewCurrencyConverter()

	for i, tc := range tests {
		t.Run(fmt.Sprintf("TestRangeFilter#%d", i), func(t *testing.T) {
			filter, err := NewRangeFilter(tc.model, tc.min, tc.max, tc.currency, converter)
			if err != nil {
				t.Errorf("cannot create filter with model regex '%s' and price range [%.2f, %.2f]: %s", tc.model, tc.min, tc.max, err)
			}

			included := filter.Include(tc.product)

			if included != tc.included {
				t.Errorf("product '%s' of price %.2f%s with model regex '%s' and range [%.2f, %.2f]: got included=%t, want included=%t", tc.product.Name, tc.product.Price, tc.product.PriceCurrency, tc.model, tc.min, tc.max, included, tc.included)
			} else {
				if included {
					t.Logf("product '%s' included by model regex '%s' and range [%.2f, %.2f]", tc.product.Name, tc.model, tc.min, tc.max)
				} else {
					t.Logf("product '%s' excluded by model regex '%s' and range [%.2f, %.2f]", tc.product.Name, tc.model, tc.min, tc.max)
				}
			}

		})
	}
}
