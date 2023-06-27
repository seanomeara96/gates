package services

import (
	"errors"
	"sort"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repositories"
)

type BundleService struct {
	bundleRepository  *repositories.BundleRepository
	productRepository *repositories.ProductRepository
}

func NewBundleService(productRepository *repositories.ProductRepository, bundleRepository *repositories.BundleRepository) *BundleService {
	return &BundleService{productRepository: productRepository, bundleRepository: bundleRepository}
}

func (s *BundleService) CreateTables() error {
	err := s.bundleRepository.CreateTables()
	if err != nil {
		return err
	}
	return nil
}

func (s *BundleService) ClearAll() error {
	err := s.bundleRepository.ClearAll()
	if err != nil {
		return err
	}
	return nil
}

func (s *BundleService) SaveRequestedBundleSize(desiredWidth float32) error {
	err := s.bundleRepository.SaveRequestedBundleSize(desiredWidth)
	if err != nil {
		return err
	}
	return nil
}

type PopularSizes = []repositories.PopularSize

func (s *BundleService) PopularSizes(limit int) (PopularSizes, error) {
	popularSizes, err := s.bundleRepository.PopularSizes(limit)
	if err != nil {
		return popularSizes, err
	}
	return popularSizes, nil
}

func (s *BundleService) BuildPressureFitBundles(limit float32) ([]models.Bundle, error) {
	var bundles []models.Bundle

	gates, err := s.productRepository.GetGates(ProductFilterParams{MaxWidth: limit})
	if err != nil {
		return bundles, err
	}
	if len(gates) < 1 {
		return bundles, nil
	}

	for _, gate := range gates {
		compatibleExtensions, err := s.productRepository.GetCompatibleExtensions(gate.Id)
		if err != nil {
			return bundles, err
		}

		bundle, err := s.BuildPressureFitBundle(limit, gate, compatibleExtensions)
		if err != nil {
			return bundles, err
		}
		bundles = append(bundles, bundle)
	}

	return bundles, nil
}

func (s *BundleService) BuildPressureFitBundle(limit float32, gate *models.Product, extensions []*models.Product) (models.Bundle, error) {
	widthLimit := limit

	var bundle models.Bundle = models.Bundle{}
	// returning a single bundle
	bundle.Qty = 1

	//  add gate to the bundle. Ensure Qty is at least 1
	if gate.Width > widthLimit {
		return bundle, errors.New("gate too big")
	}

	if gate.Qty < 1 {
		gate.Qty = 1
	}

	bundle.Gates = append(bundle.Gates, *gate)

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
		for ii := 0; ii < len(bundle.Extensions); ii++ {
			var bundleExtension *models.Product = &bundle.Extensions[ii]

			if bundleExtension.Id == extension.Id {
				existingExtension = bundleExtension
			}
		}

		if existingExtension != nil {
			existingExtension.Qty++
			widthLimit -= existingExtension.Width
		} else {
			extension.Qty = 1
			bundle.Extensions = append(bundle.Extensions, *extension)
			widthLimit -= extension.Width
		}
	}

	bundle.ComputeMetaData()
	return bundle, nil
}

func (s *BundleService) SaveBundle(bundle models.Bundle) (int64, error) {
	bundleId, err := s.bundleRepository.SaveBundleAsProduct(bundle.Product)
	if err != nil {
		return 0, err
	}
	for iii := 0; iii < len(bundle.Gates); iii++ {
		gate := bundle.Gates[iii]
		err = s.bundleRepository.SaveBundleGate(gate.Id, bundleId, gate.Qty)
		if err != nil {
			return 0, err
		}
	}
	for ii := 0; ii < len(bundle.Extensions); ii++ {
		extension := bundle.Extensions[ii]
		err := s.bundleRepository.SaveBundleExtension(extension.Id, bundleId, extension.Qty)
		if err != nil {
			return 0, err
		}
	}
	return bundleId, nil
}
