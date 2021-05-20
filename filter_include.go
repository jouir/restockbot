package main

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

// IncludeFilter struct to store the compiled regex used for the inclusion
type IncludeFilter struct {
	regex *regexp.Regexp
}

// NewIncludeFilter to create an IncludeFilter
func NewIncludeFilter(regex string) (*IncludeFilter, error) {
	var err error
	var compiledRegex *regexp.Regexp

	log.Debugf("compiling include filter regex")
	if regex != "" {
		compiledRegex, err = regexp.Compile(regex)
		if err != nil {
			return nil, err
		}
	}

	return &IncludeFilter{regex: compiledRegex}, nil
}

// Include returns true when the product name matches the regex
// implements the Filter interface
func (f *IncludeFilter) Include(product *Product) bool {
	if f.regex == nil {
		return true
	}
	if f.regex.MatchString(product.Name) {
		log.Debugf("product %s included because it matches the include regex", product.Name)
		return true
	}
	log.Debugf("product %s excluded because it doesn't match the include regex", product.Name)
	return false
}
