package main

import (
	"database/sql"
	"testing"

	"github.com/seanomeara96/gates/repositories"
	"github.com/seanomeara96/gates/services"

	_ "github.com/mattn/go-sqlite3"
)

func TestBuildPressureFitBundles(t *testing.T) {
	db, err := sql.Open("sqlite3", "../main.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	productRepo := repositories.NewProductRepository(db)
	bundleRepo := repositories.NewBundleRepository(db)
	bundleService := services.NewBundleService(productRepo, bundleRepo)
	bundles, err := bundleService.BuildPressureFitBundles(125)
	if err != nil {
		t.Error(err)
	}
	if len(bundles) < 1 {
		t.Error("expected at least 1 bundle")
	}

}
