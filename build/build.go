package build

import (
	"sort"

	"github.com/seanomeara96/gates/components"
)

func Bundle(gate components.Gate, extensions components.Extensions, desiredWidth float32) components.Bundle {
	var bundle components.Bundle
	bundle.Gate = gate

	sort.Slice(extensions, func(i int, j int) bool {
		return extensions[i].Width > extensions[j].Width
	})

	smallestExtension := extensions[len(extensions)-1]

	widthRemaining := desiredWidth - gate.Width

	index := 0
	for widthRemaining > 0 {
		if index > len(extensions) {
			break
		}

		currentExtension := extensions[index]
		currentExtension.Qty = 1

		if widthRemaining >= currentExtension.Width || (widthRemaining < smallestExtension.Width && widthRemaining > 0) {
			var matchingExtension *components.Extension
			// find matching extension
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
		}

	}

	bundle.MaxLength = bundle.Gate.Width
	bundle.Price = bundle.Gate.Price
	for _, extension := range bundle.Extensions {
		bundle.MaxLength += extension.Width * float32(extension.Qty)
		bundle.Price += extension.Price * float32(extension.Qty)
	}

	// need to test if the width of the bundle minus the threshold actuall accomodates the requested size
	// if it doesnt set the qty of the bundle to 0
	bundle.Qty = 1
	return bundle
}
