package main

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

// DefaultCurrency to fallback when no currency is provided
const DefaultCurrency = "USD"

// RangeFilter to store the pattern to match product model and price limits
type RangeFilter struct {
	model     *regexp.Regexp
	min       float64
	max       float64
	currency  string
	converter *CurrencyConverter
}

// NewRangeFilter to create a RangeFilter
func NewRangeFilter(regex string, min float64, max float64, currency string, converter *CurrencyConverter) (*RangeFilter, error) {
	var err error
	var compiledRegex *regexp.Regexp

	log.Debugf("compiling model filter regex")
	if regex != "" {
		compiledRegex, err = regexp.Compile(regex)
		if err != nil {
			return nil, err
		}
	}

	var detectedCurrency string
	if currency != "" {
		detectedCurrency = currency
	} else {
		detectedCurrency = DefaultCurrency
	}

	return &RangeFilter{
		model:     compiledRegex,
		min:       min,
		max:       max,
		converter: converter,
		currency:  detectedCurrency,
	}, nil
}

// Include returns false when a product name matches the model regex and price is outside of the range
// implements the Filter interface
func (f *RangeFilter) Include(product *Product) bool {
	// include products with a missing model regex
	if f.model == nil {
		log.Debugf("product %s included because range filter model is missing", product.Name)
		return true
	}

	// include products with a different model
	if !f.model.MatchString(product.Name) {
		log.Debugf("product %s included because range filter model hasn't been detected in the product name", product.Name)
		return true
	}

	// convert price to the filter currency
	convertedPrice, err := f.converter.Convert(product.Price, product.PriceCurrency, f.currency)
	if err != nil {
		log.Warnf("could not convert price %.2f %s to %s for range filter: %s", product.Price, product.PriceCurrency, f.currency, err)
		return true
	}

	// include prices with unlimited maximum if min is respected
	if f.max == 0 && convertedPrice > f.max && f.min <= convertedPrice {
		log.Debugf("product %s included because max value is unlimited and converted price of %.2f%s is higher than lower limit of %.2f%s", product.Name, convertedPrice, f.currency, f.min, f.currency)
		return true
	}

	// include prices inside the range
	if f.min <= convertedPrice && convertedPrice <= f.max {
		log.Debugf("product %s included because range filter model matches and converted price is inside of the range", product.Name)
		return true
	}

	log.Debugf("product %s excluded because range filter model matches and converted price is outside of the range", product.Name)
	return false
}
