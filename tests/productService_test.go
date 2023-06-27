package main

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/repositories"
	"github.com/seanomeara96/gates/services"
)

func TestGetProductById(t *testing.T) {
	db, err := sql.Open("sqlite3", "../main.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)

	product, err := productService.GetProductByID(2)
	if err != nil {
		t.Error(err)
	}

	if product.Id != 2 {
		t.Errorf("Expected product with an ID of %d, got%d instead \n", 2, 2)
	}

}

func TestGetGatesService(t *testing.T) {
	db, err := sql.Open("sqlite3", "../main.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)

	gates, err := productService.GetGates(services.ProductFilterParams{})
	if err != nil {
		t.Error(err)
	}
	if len(gates) < 1 {
		t.Errorf("expected more than %d gates", len(gates))
	}
}

func TestGetExtensions(t *testing.T) {
	db, err := sql.Open("sqlite3", "../main.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	extensions, err := productService.GetExtensions(services.ProductFilterParams{})
	if err != nil {
		t.Error(err)
	}
	if len(extensions) < 1 {
		t.Errorf("expected more than %d extensions", len(extensions))
	}
}
