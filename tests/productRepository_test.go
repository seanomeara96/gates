package main

import (
	"testing"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repositories"
)

func TestProductFields(t *testing.T) {

	product := models.Product{
		Id:        1,
		Type:      "gate",
		Name:      "product 1",
		Width:     1.1,
		Price:     1.1,
		Img:       "path/to/img",
		Color:     "white",
		Tolerance: 1.1,
	}

	repositories.ProductFields(product)

}
