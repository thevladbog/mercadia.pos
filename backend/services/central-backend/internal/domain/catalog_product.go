package domain

import (
	"errors"
	"time"
)

var ErrInvalidCatalogProductInput = errors.New("invalid catalog product input")

type CatalogProduct struct {
	ID             string
	StoreID        string
	Name           string
	Barcodes       []string
	UnitPriceMinor int64
	TaxCategoryID  string
	Active         bool
	Version        int64
	UpdatedAt      time.Time
}

func NewCatalogProduct(product CatalogProduct) (CatalogProduct, error) {
	if product.ID == "" || product.StoreID == "" || product.Name == "" || len(product.Barcodes) == 0 || product.UnitPriceMinor < 0 {
		return CatalogProduct{}, ErrInvalidCatalogProductInput
	}
	product.Barcodes = append([]string(nil), product.Barcodes...)
	if product.Version <= 0 {
		product.Version = 1
	}
	return product, nil
}

func (p CatalogProduct) PrimaryBarcode() string {
	if len(p.Barcodes) == 0 {
		return ""
	}
	return p.Barcodes[0]
}
