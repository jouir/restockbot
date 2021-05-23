package main

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestCurrencyConverterConvert(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://cdn.jsdelivr.net/gh/fawazahmed0/currency-api@1/latest/currencies/eur/chf.json",
		httpmock.NewStringResponder(200, `{"date": "2021-05-22", "chf": 1.093894}`))

	tests := []struct {
		amount      float64
		source      string
		destination string
		expected    float64
	}{
		{1.0, "EUR", "EUR", 1.0},      // same currency (EUR/EUR)
		{1.0, "EUR", "CHF", 1.093894}, // different currency (EUR/CHF)
		{1.0, "EUR", "CHF", 1.093894}, // different currency (EUR/CHF) with cache
	}

	converter := NewCurrencyConverter()

	for i, tc := range tests {
		t.Run(fmt.Sprintf("TestCurrencyConverterConvert#%d", i), func(t *testing.T) {

			converted, err := converter.Convert(tc.amount, tc.source, tc.destination)

			if err != nil {
				t.Errorf("could not convert %.2f from %s to %s: %s", tc.amount, tc.source, tc.destination, err)
			} else if converted != tc.expected {
				t.Errorf("to convert %.2f from %s to %s, got '%.2f', want '%.2f'", tc.amount, tc.source, tc.destination, converted, tc.expected)
			} else {
				t.Logf("to convert %.2f from %s to %s, got '%.2f', want '%.2f'", tc.amount, tc.source, tc.destination, converted, tc.expected)
			}
		})
	}
}
