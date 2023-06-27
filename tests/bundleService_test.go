package tests

import (
	"database/sql"
	"testing"

	"github.com/seanomeara96/gates/models"
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

func TestSaveBundle(t *testing.T) {
	db, err := sql.Open("sqlite3", "../main.db")
	if err != nil {
		t.Error(err)
		return
	}

	bundleRepo := repositories.NewBundleRepository(db)
	productRepo := repositories.NewProductRepository(db)
	bundleService := services.NewBundleService(productRepo, bundleRepo)
	err = bundleService.CreateTables()
	if err != nil {
		t.Error(err)
		return
	}
	err = bundleService.ClearAll()
	if err != nil {
		t.Error(err)
		return
	}
	var bundle models.Bundle = models.Bundle{}
	gate := models.Product{
		Id:        1,
		Type:      "gate",
		Name:      "Gate",
		Width:     100,
		Price:     1.1,
		Img:       "img",
		Color:     "black",
		Tolerance: 1.1,
		Qty:       1,
	}
	bundle.Gates = append(bundle.Gates, gate)

	extension := models.Product{
		Id:        3,
		Type:      "extension",
		Name:      "Extension",
		Width:     30,
		Price:     1.1,
		Img:       "img",
		Color:     "black",
		Tolerance: 1.1,
		Qty:       1,
	}
	bundle.Extensions = append(bundle.Extensions, extension)

	bundle.ComputeMetaData()

	_, err = bundleService.SaveBundle(bundle)
	if err != nil {
		t.Error(err)
	}

}
