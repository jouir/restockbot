package main

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

// Parser interface to parse an external service and return a list of products
type Parser interface {
	Parse() ([]*Product, error)
	String() string
}

// filterInclusive returns a list of products matching the include regex
func filterInclusive(includeRegex *regexp.Regexp, products []*Product) []*Product {
	var filtered []*Product
	if includeRegex != nil {
		for _, product := range products {
			if includeRegex.MatchString(product.Name) {
				log.Debugf("product %s included because it matches the include regex", product.Name)
				filtered = append(filtered, product)
			} else {
				log.Debugf("product %s excluded because it does not match the include regex", product.Name)
			}
		}
		return filtered
	}
	return products
}

// filterExclusive returns a list of products that don't match the exclude regex
func filterExclusive(excludeRegex *regexp.Regexp, products []*Product) []*Product {
	var filtered []*Product
	if excludeRegex != nil {
		for _, product := range products {
			if excludeRegex.MatchString(product.Name) {
				log.Debugf("product %s excluded because it matches the exclude regex", product.Name)
			} else {
				log.Debugf("product %s included because it does not match the exclude regex", product.Name)
				filtered = append(filtered, product)
			}
		}
		return filtered
	}
	return products
}
