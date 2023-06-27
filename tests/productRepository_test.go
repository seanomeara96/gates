package tests

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/repositories"
)

func TestGetCompatibleExtensions(t *testing.T) {
	db, err := sql.Open("sqlite3", "../main.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	productRepo := repositories.NewProductRepository(db)
	extensions, err := productRepo.GetCompatibleExtensions(1)
	if err != nil {
		t.Error(err)
	}
	if len(extensions) != 3 {
		t.Error("expected 3 extensions")
	}
}

func TestMaxWidthFilter(t *testing.T) {
	db, err := sql.Open("sqlite3", "../main.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	productRepo := repositories.NewProductRepository(db)
	filters := repositories.ProductFilterParams{
		MaxWidth: 35,
	}
	extensions, err := productRepo.GetExtensions(filters)
	if err != nil {
		t.Error(err)
	}
	if len(extensions) > 4 {
		t.Error("expected 4 extensions")
	}
}

func TestNoGates(t *testing.T) {
	db, err := sql.Open("sqlite3", "../main.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	productRepo := repositories.NewProductRepository(db)
	filters := repositories.ProductFilterParams{
		MaxWidth: 70,
	}
	gates, err := productRepo.GetGates(filters)
	if err != nil {
		t.Error(err)
	}

	if len(gates) > 0 {
		t.Error("expected no gates")
	}

}
