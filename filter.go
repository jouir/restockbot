package main

// Filter interface to include a product based on filters
type Filter interface {
	Include(*Product) bool
}
