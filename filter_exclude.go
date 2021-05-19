package main

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

type ExcludeFilter struct {
	regex *regexp.Regexp
}

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

// Filter excludes product with name matching the regex
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
