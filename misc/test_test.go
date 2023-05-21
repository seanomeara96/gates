package main

import (
	"fmt"
	"testing"

	"github.com/seanomeara96/gates/types"
)

func TestBuildPressureFitBundle(t *testing.T) {
	// Initialize two Gate structs
	blackGate := types.Gate{
		Product: types.Product{
			Id:    1,
			Name:  "Black Gate",
			Width: 80,
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
			Width: 80,
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

	bundle, err := buildPressureFitBundle(100, blackGate, blackGateExtensions)

	// limitof 100
	// gate should be added with 20cm leftover
	// one 15 cm extension and on 5 cm extension

	if err != nil {
		t.Error("there was an error")
	}

	if len(bundle.Extensions) != 2 {
		t.Errorf("incorrect number of extensions, expected 2 got %d", len(bundle.Extensions))
	}

	for ii := 0; ii < len(bundle.Extensions); ii++ {
		if bundle.Extensions[ii].Width != float32(5) || bundle.Extensions[ii].Width != float32(15) {
			t.Error("incorrect widths")
		}
	}

}
