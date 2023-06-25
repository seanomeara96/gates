package main

import (
	"database/sql"
	"testing"

	"github.com/seanomeara96/gates/repositories"
	"github.com/seanomeara96/gates/services"

	_ "github.com/mattn/go-sqlite3"
)

func loadDB() *sql.DB {
	db, err := sql.Open("sqlite3", "../main.db")
	if err != nil {
		panic(err)
	}
	return db
}

func TestBuildPressureFitBundles(t *testing.T) {
	db := loadDB()
	productRepo := repositories.NewProductRepository(db)
	bundleService := services.NewBundleService(productRepo)
	bundles, err := bundleService.BuildPressureFitBundles(125)
	if err != nil {
		t.Error(err)
	}
	if len(bundles) < 1 {
		t.Error("expected at least 1 bundle")
	}

}
