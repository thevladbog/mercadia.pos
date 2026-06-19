package domain

import "errors"

var ErrInvalidProductInput = errors.New("invalid product input")

type Product struct {
	ID             string
	Name           string
	Barcodes       []string
	UnitPriceMinor int64
	TaxCategoryID  string
	Active         bool
}

func NewProduct(product Product) (Product, error) {
	if product.ID == "" || product.Name == "" || len(product.Barcodes) == 0 || product.UnitPriceMinor < 0 {
		return Product{}, ErrInvalidProductInput
	}
	product.Active = true
	product.Barcodes = append([]string(nil), product.Barcodes...)
	return product, nil
}

func (p Product) PrimaryBarcode() string {
	if len(p.Barcodes) == 0 {
		return ""
	}
	return p.Barcodes[0]
}
