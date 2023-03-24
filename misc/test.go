package main

import (
	"fmt"
	"sort"
)

type Gate struct {
	Id        int
	Name      string
	Width     float32
	Price     float32
	Img       string
	Tolerance float32
	Color     string
	Qty       int
}
type Gates []Gate
type Extension struct {
	Id    int
	Name  string
	Width float32
	Price float32
	Img   string
	Color string
	Qty   int
}
type Extensions []Extension
type Bundle struct {
	Gate       Gate
	Extensions Extensions
}

func buildBundle(gate Gate, extensions Extensions, desiredWidth float32) Bundle {
	var bundle Bundle
	bundle.Gate = gate

	sort.Slice(extensions, func(i int, j int) bool {
		return extensions[i].Width < extensions[j].Width
	})

	smallestExtension := extensions[len(extensions)-1]

	widthRemaining := desiredWidth - gate.Width

	index := 0
	for widthRemaining > 0 {
		if index > len(extensions) {
			break
		}

		currentExtension := extensions[index]

		if currentExtension.Width <= widthRemaining {
			var matchingExtension *Extension
			for i := 0; i < len(bundle.Extensions); i++ {
				if currentExtension.Id == bundle.Extensions[i].Id {
					matchingExtension = &bundle.Extensions[i]
				}
			}

			if matchingExtension != nil {
				matchingExtension.Qty++
			} else {
				bundle.Extensions = append(bundle.Extensions, currentExtension)
			}

			widthRemaining = widthRemaining - currentExtension.Width
			continue
		}

		if currentExtension.Width > widthRemaining && currentExtension.Id != smallestExtension.Id {
			index++
			continue
		}

		if widthRemaining < smallestExtension.Width && widthRemaining > 0 {
			bundle.Extensions = append(bundle.Extensions, smallestExtension)
			widthRemaining = widthRemaining - smallestExtension.Width
		}

	}

	return bundle
}

func main() {
	// Create some gate objects
	g1 := Gate{Id: 1, Name: "My Gate", Width: 30.0, Price: 100.0, Img: "gate1.jpg", Tolerance: 1.0, Color: "White", Qty: 1}

	// Create some extension objects
	e1 := Extension{Id: 1, Name: "Extension 1", Width: 30.0, Price: 50.0, Img: "ext1.jpg", Color: "White", Qty: 1}
	e2 := Extension{Id: 2, Name: "Extension 2", Width: 35.0, Price: 60.0, Img: "ext2.jpg", Color: "Black", Qty: 1}
	e3 := Extension{Id: 3, Name: "Extension 3", Width: 40.0, Price: 70.0, Img: "ext3.jpg", Color: "Brown", Qty: 1}

	// Create extensions slice and append extension objects
	extensions := Extensions{}
	extensions = append(extensions, e1)
	extensions = append(extensions, e2)
	extensions = append(extensions, e3)

	bundle := buildBundle(g1, extensions, 200)

	fmt.Println(bundle)

	//var sumWidth = bundle.Gate.Width
	// TODO sum the total width of the bundle and test the threshold of the gate to see if it will actually fit
	// this algorithm is only reliable if the smallest extension width == the threshold of the gate
}
