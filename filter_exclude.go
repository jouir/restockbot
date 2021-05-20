package main

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

// ExcludeFilter struct to store the compiled regex used for the exclusion
type ExcludeFilter struct {
	regex *regexp.Regexp
}

// NewExcludeFilter to create an ExcludeFilter
func NewExcludeFilter(regex string) (*ExcludeFilter, error) {
	var err error
	var compiledRegex *regexp.Regexp

	log.Debugf("compiling exclude filter regex")
	if regex != "" {
		compiledRegex, err = regexp.Compile(regex)
		if err != nil {
			return nil, err
		}
	}

	return &ExcludeFilter{regex: compiledRegex}, nil
}

// Include returns false when the product name matches the regex
// implements the Filter interface
func (f *ExcludeFilter) Include(product *Product) bool {
	if f.regex == nil {
		return true
	}
	if f.regex.MatchString(product.Name) {
		log.Debugf("product %s excluded because it matches the exclude regex", product.Name)
		return false
	}
	log.Debugf("product %s included because it doesn't match the exclude regex", product.Name)
	return true
}
