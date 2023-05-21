package main

import (
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/seanomeara96/gates/types"
)

func main() {
	// Initialize two Gate structs
	blackGate := types.Gate{
		Product: types.Product{
			Id:    1,
			Name:  "Black Gate",
			Width: 10.5,
			Price: 19.99,
			Img:   "black_gate.jpg",
			Color: "Black",
			Qty:   1,
		},
		Tolerance: 5,
	}

	whiteGate := types.Gate{
		Product: types.Product{
			Id:    2,
			Name:  "White Gate",
			Width: 12.0,
			Price: 24.99,
			Img:   "white_gate.jpg",
			Color: "White",
			Qty:   1,
		},
		Tolerance: 5,
	}

	// Initialize three Extension structs for each Gate
	blackGateExtensions := make([]types.Extension, 3)
	whiteGateExtensions := make([]types.Extension, 3)

	for i := 0; i < 3; i++ {
		blackGateExtensions[i] = types.Extension{
			Product: types.Product{
				Id:    i + 1,
				Name:  fmt.Sprintf("Black Extension %d", i+1),
				Width: 5.0 * float32(i+1),
				Price: 9.99,
				Img:   fmt.Sprintf("black_extension%d.jpg", i+1),
				Color: "Black",
				Qty:   1,
			},
		}

		whiteGateExtensions[i] = types.Extension{
			Product: types.Product{
				Id:    i + 1,
				Name:  fmt.Sprintf("White Extension %d", i+1),
				Width: 5.0 * float32(i+1),
				Price: 11.99,
				Img:   fmt.Sprintf("white_extension%d.jpg", i+1),
				Color: "White",
				Qty:   1,
			},
		}
	}

	var gates types.Gates = make(types.Gates, 2)
	gates[0] = blackGate
	gates[1] = whiteGate

	var extensions []types.Extension
	for i := 0; i < 3; i++ {
		extensions = append(extensions, blackGateExtensions[i])
		extensions = append(extensions, whiteGateExtensions[i])
	}

	out, err := buildPressureFitBundle(100, blackGate, blackGateExtensions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(out)
}

func buildPressureFitBundle(limit int, gate types.Gate, extensions types.Extensions) (types.Bundle, error) {
	widthLimit := float32(limit)

	var bundle types.Bundle = types.Bundle{}
	// returning a single bundle
	bundle.Qty = 1

	//  add gate to the bundle. Ensure Qty is at least 1
	if gate.Width > widthLimit {
		return bundle, errors.New("Gate too big.")
	}

	if gate.Qty < 1 {
		gate.Qty = 1
	}

	bundle.Gates = append(bundle.Gates, gate)

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
		if extension.Width > widthLimit && override == false {
			//  extension too big, try next extension size down
			extensionIndex++
			continue
		}

		// check if extension already exists in the bundle and if so, increment the qty, else add it with a qty of 1
		var existingExtension *types.Extension
		for ii := 0; ii < len(bundle.Extensions); ii++ {
			bundleExtension := bundle.Extensions[ii]

			if bundleExtension.Id == extension.Id {
				existingExtension = &bundleExtension
			}
		}

		if existingExtension != nil {
			existingExtension.Qty++
			widthLimit -= existingExtension.Width
		} else {
			extension.Qty = 1
			bundle.Extensions = append(bundle.Extensions, extension)
			widthLimit -= extension.Width
		}
	}

	bundle.ComputeMetaData()
	return bundle, nil
}
