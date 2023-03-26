package build

import (
	"testing"

	"github.com/seanomeara96/gates/components"
)

func TestBuild(t *testing.T) {
	// Create some gate objects
	g1 := components.Gate{Id: 1, Name: "My Gate", Width: 30.0, Price: 100.0, Img: "gate1.jpg", Tolerance: 1.0, Color: "White", Qty: 1}

	// Create some extension objects
	e1 := components.Extension{Id: 1, Name: "Extension 1", Width: 30.0, Price: 50.0, Img: "ext1.jpg", Color: "White", Qty: 1}
	e2 := components.Extension{Id: 2, Name: "Extension 2", Width: 35.0, Price: 60.0, Img: "ext2.jpg", Color: "Black", Qty: 1}
	e3 := components.Extension{Id: 3, Name: "Extension 3", Width: 40.0, Price: 70.0, Img: "ext3.jpg", Color: "Brown", Qty: 1}

	// Create extensions slice and append extension objects
	extensions := components.Extensions{}
	extensions = append(extensions, e1)
	extensions = append(extensions, e2)
	extensions = append(extensions, e3)

	bundle := Bundle(g1, extensions, 200)

	if bundle.Gate.Qty != 1 {
		t.Errorf("Exepected gate qty of 1, got %d instead", bundle.Gate.Qty)
	}

	if bundle.MaxLength == 0 {
		t.Errorf("Expected bundle to have a max length greater than zero, got %.2f", bundle.MaxLength)
	}

	if bundle.Price == 0 {
		t.Errorf("Expected bundle to have a price greater than zero, got %.2f", bundle.Price)
	}

}
