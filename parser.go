package main

// Parser interface to parse an external service and return a list of products
type Parser interface {
	Parse() ([]*Product, error)
	String() string
	ShopName() (string, error)
}
