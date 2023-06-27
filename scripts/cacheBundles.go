package scripts

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"reflect"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repositories"
	"github.com/seanomeara96/gates/services"
)

func cache() {

	db, err := sql.Open("sqlite3", filepath.Join("/home/user/gates/scripts", "../main.db"))
	if err != nil {
		log.Fatal(err)
	}

	bundleRepo := repositories.NewBundleRepository(db)
	productRepo := repositories.NewProductRepository(db)

	bundleService := services.NewBundleService(productRepo, bundleRepo)
	productService := services.NewProductService(productRepo)

	err = bundleService.CreateTables()
	if err != nil {
		log.Fatal(err)
	}

	err = bundleService.ClearAll()
	if err != nil {
		log.Fatal(err)
	}

	popularSizes, err := bundleService.PopularSizes(3)
	if err != nil {
		log.Fatal(err)
	}

	var bundles []models.Bundle

	// for each common size request build a bundle and save it to the database
	for i := 0; i < len(popularSizes); i++ {

		desiredWidth := popularSizes[i].Size
		fmt.Println("desiredWidth", desiredWidth)

		gates, err := productService.GetGates(services.ProductFilterParams{MaxWidth: desiredWidth})
		if err != nil {
			log.Fatal(err)
		}

		// no gates matched the request
		if len(gates) < 1 {
			continue
		}

		for i := 0; i < len(gates); i++ {
			gate := gates[i]

			extensions, err := productService.GetCompatibleExtensions(gate.Id)
			if err != nil {
				log.Fatal(err)
			}

			bundle, err := bundleService.BuildPressureFitBundle(desiredWidth, gate, extensions)
			if err != nil {
				log.Print("error building bundle")
				continue
			}
			bundles = append(bundles, bundle)
		}
	}

	// filter duplicate bundles && save unique bundles to the database
	var uniqueBundles []models.Bundle
	for i := 0; i < len(bundles); i++ {
		bundle := bundles[i]
		encountered := false
		for _, value := range uniqueBundles {
			if reflect.DeepEqual(bundle, value) {
				encountered = true
			}
		}
		if !encountered {
			uniqueBundles = append(uniqueBundles, bundle)
		}
	}

	for i := 0; i < len(uniqueBundles); i++ {
		bundleService.SaveBundle(uniqueBundles[i])
	}
}
