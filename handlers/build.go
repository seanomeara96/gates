package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repos"
	"github.com/seanomeara96/gates/repos/cache"
)

func BuildPressureFitBundles(products *cache.CachedProductRepo, limit float32) ([]models.Bundle, error) {
	var bundles []models.Bundle

	gates, err := products.GetProducts(repos.ProductFilterParams{MaxWidth: limit, Type: models.ProductTypeGate})
	if err != nil {
		return bundles, err
	}
	if len(gates) < 1 {
		return bundles, nil
	}

	for _, gate := range gates {
		compatibleExtensions, err := products.GetCompatibleExtensionsByGateID(gate.Id)
		if err != nil {
			return bundles, err
		}

		bundle, err := BuildPressureFitBundle(limit, gate, compatibleExtensions)
		if err != nil {
			return bundles, err
		}
		bundles = append(bundles, bundle)
	}

	return bundles, nil
}

func BuildPressureFitBundle(limit float32, gate *models.Product, extensions []*models.Product) (models.Bundle, error) {
	widthLimit := limit

	var bundle = models.Bundle{}

	// returning a single bundle
	bundle.Qty = 1

	//  add gate to the bundle. Ensure Qty is at least 1
	if gate.Width > widthLimit {
		return bundle, errors.New("gate too big")
	}

	if gate.Qty < 1 {
		gate.Qty = 1
	}

	bundle.Components = append(bundle.Components, *gate)

	widthLimit -= gate.Width

	// sort extensions to ensure width descending
	sort.Slice(extensions, func(i int, j int) bool {
		return extensions[i].Width > extensions[j].Width
	})

	extensionIndex := 0
	for widthLimit > 0 {

		// we want to add one more extension if the width remaining > 0 but we've reached the last extension
		var override bool = false
		if extensionIndex >= len(extensions) {
			extensionIndex--
			override = true
		}

		extension := extensions[extensionIndex]
		if extension.Width > widthLimit && !override {
			//  extension too big, try next extension size down
			extensionIndex++
			continue
		}

		// check if extension already exists in the bundle and if so, increment the qty, else add it with a qty of 1
		var existingExtension *models.Product
		for ii := 1; ii < len(bundle.Components); ii++ {
			var bundleExtension *models.Product = &bundle.Components[ii]

			if bundleExtension.Id == extension.Id {
				existingExtension = bundleExtension
			}
		}

		if existingExtension != nil {
			existingExtension.Qty++
			widthLimit -= existingExtension.Width
		} else {
			extension.Qty = 1
			bundle.Components = append(bundle.Components, *extension)
			widthLimit -= extension.Width
		}
	}
	bundle.ComputeMetaData()
	return bundle, nil
}

func SaveRequestedBundleSize(db *sql.DB, desiredWidth float32) error {
	_, err := db.Exec("INSERT INTO bundle_sizes (type, size) VALUES ('pressure fit', ?)", desiredWidth)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) BuildBundle(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("build endpoint: failed to parse form: %w", err)
	}

	desiredWidth, err := strconv.ParseFloat(r.Form.Get("desired-width"), 32)
	if err != nil {
		return fmt.Errorf("build endpoint: failed to parse desired width: %w", err)
	}

	/*
		static max witdth for now
		could consider adding a max width to the gates in the DB and enforcing the
		constraint from there
	*/
	maxWidth := 220.00
	if desiredWidth > maxWidth {
		desiredWidth = maxWidth
	}

	if err := SaveRequestedBundleSize(h.db, float32(desiredWidth)); err != nil {
		return fmt.Errorf("build endpoint: failed to save requested bundle size: %w", err)
	}

	bundles, err := BuildPressureFitBundles(h.productCache, float32(desiredWidth))
	if err != nil {
		return fmt.Errorf("build endpoint: failed to build pressure fit bundles: %w", err)
	}

	data := map[string]any{
		"RequestedBundleSize": float32(desiredWidth),
		"Bundles":             bundles,
		"Env":                 h.cfg.Mode,
	}

	return h.rndr.Page(w, "build-results", data)
}
