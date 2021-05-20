package main

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

// RangeFilter to store the pattern to match product model and price limits
type RangeFilter struct {
	model *regexp.Regexp
	min   float64
	max   float64
}

// NewRangeFilter to create a RangeFilter
func NewRangeFilter(regex string, min float64, max float64) (*RangeFilter, error) {
	var err error
	var compiledRegex *regexp.Regexp

	log.Debugf("compiling model filter regex")
	if regex != "" {
		compiledRegex, err = regexp.Compile(regex)
		if err != nil {
			return nil, err
		}
	}

	return &RangeFilter{
		model: compiledRegex,
		min:   min,
		max:   max,
	}, nil
}

// Include returns false when a product name matches the model regex and price is outside of the range
// implements the Filter interface
func (f *RangeFilter) Include(product *Product) bool {
	if f.model == nil {
		return true
	}
	if f.model.MatchString(product.Name) && product.Price < f.min || product.Price > f.max {
		log.Debugf("product %s excluded because price for the model is outside of the range [%.2f-%.2f]", product.Name, f.min, f.max)
		return false
	}
	log.Debugf("product %s included because price range filter is not applicable", product.Name)
	return true
}
