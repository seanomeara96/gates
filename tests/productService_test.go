package main

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/repositories"
	"github.com/seanomeara96/gates/services"
)

func TestGetProductById(t *testing.T) {
	db, err := sql.Open("sqlite3", "main.db")
	if err != nil {
		t.Error(err)
	}
	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)

	product, err := productService.GetProductById(2)
	if err != nil {
		t.Error(err)
	}

}
