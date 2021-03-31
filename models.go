package main

import (
	"gorm.io/gorm"
)

// Product is self-explainatory
type Product struct {
	gorm.Model
	Name          string  `gorm:"not null" json:"name"`
	URL           string  `gorm:"unique" json:"url"`
	Price         float64 `gorm:"not null" json:"price"`
	PriceCurrency string  `gorm:"not null" json:"price_currency"`
	Available     bool    `gorm:"not null;default:false" json:"available"`
	ShopID        uint    `json:"shop_id"`
	Shop          Shop    `json:"shop"`
}

// Equal compares a database product to another product
func (p *Product) Equal(other *Product) bool {
	return p.URL == other.URL && p.Available == other.Available
}

// IsValid returns true when a Product has all required values
func (p *Product) IsValid() bool {
	if p.Name == "" || p.URL == "" {
		return false
	}
	if p.Available && p.PriceCurrency == "" {
		return false
	}
	return true
}

// Merge one product with another
func (p *Product) Merge(o *Product) {
	p.Price = o.Price
	p.PriceCurrency = o.PriceCurrency
	p.Available = o.Available
}

// ToMerge detects if a product needs to be merged with another one
func (p *Product) ToMerge(o *Product) bool {
	return p.Price != o.Price || p.PriceCurrency != o.PriceCurrency || p.Available != o.Available
}

// Shop represents a retailer website
type Shop struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"unique" json:"name"`
}
