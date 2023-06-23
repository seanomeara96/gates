package main

import (
	"fmt"
	"testing"
)

func TestBuildPressureFitBundle(t *testing.T) {
	// Initialize two Gate structs
	blackGate := Gate{
		Product: Product{
			Id:    1,
			Name:  "Black Gate",
			Width: 76,
			Price: 19.99,
			Img:   "black_gate.jpg",
			Color: "Black",
			Qty:   1,
		},
		Tolerance: 6,
	}

	whiteGate := Gate{
		Product: Product{
			Id:    2,
			Name:  "White Gate",
			Width: 76,
			Price: 24.99,
			Img:   "white_gate.jpg",
			Color: "White",
			Qty:   1,
		},
		Tolerance: 6,
	}

	// Initialize three Extension structs for each Gate
	blackGateExtensions := make([]Extension, 3)
	whiteGateExtensions := make([]Extension, 3)

	for i := 0; i < 3; i++ {
		width := 7
		if i == 1 {
			width = 32
		}
		if i == 2 {
			width = 64
		}
		blackGateExtensions[i] = Extension{
			Product: Product{
				Id:    i + 1,
				Name:  fmt.Sprintf("Black Extension %d", i+1),
				Width: float32(width),
				Price: 9.99,
				Img:   fmt.Sprintf("black_extension%d.jpg", i+1),
				Color: "Black",
				Qty:   1,
			},
		}

		whiteGateExtensions[i] = Extension{
			Product: Product{
				Id:    i + 1,
				Name:  fmt.Sprintf("White Extension %d", i+1),
				Width: float32(width),
				Price: 11.99,
				Img:   fmt.Sprintf("white_extension%d.jpg", i+1),
				Color: "White",
				Qty:   1,
			},
		}
	}

	var gates Gates = make(Gates, 2)
	gates[0] = blackGate
	gates[1] = whiteGate

	desiredWidth := 125

	bundle, err := BuildPressureFitBundle(float32(desiredWidth), blackGate, blackGateExtensions)

	// limitof 100
	// gate should be added with 20cm leftover
	// one 15 cm extension and on 5 cm extension

	if err != nil {
		t.Error("there was an error")
	}

	//if len(bundle.Extensions) != 2 {
	//	t.Errorf("incorrect number of extensions, expected 2 got %d", len(bundle.Extensions))
	//}

	for ii := 0; ii < len(bundle.Extensions); ii++ {
		if bundle.Extensions[ii].Width != float32(5) {
			continue
		}

		if bundle.Extensions[ii].Width != float32(15) {
			continue
		}

		t.Error("incorrect widths")
	}

	// bundle.Width > desiredWidth - tolerance && bundle.Width < desiredWidth + tolerance

	isBundleWithinTolerance := bundle.Width > float32(desiredWidth)-bundle.Tolerance && bundle.Width < float32(desiredWidth)+bundle.Tolerance

	if !isBundleWithinTolerance {
		t.Errorf("incorrect bundle size, expected %f but got %f", float32(desiredWidth), bundle.Width)
	}

}
