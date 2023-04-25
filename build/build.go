package build

import (
	"sort"

	"github.com/seanomeara96/gates/types"
)

func Bundle(gate types.Gate, extensions types.Extensions, desiredWidth float32) types.Bundle {
	var bundle types.Bundle
	if gate.Qty < 1 {
		gate.Qty = 1
	}
	bundle.Gates = append(bundle.Gates, gate)

	sort.Slice(extensions, func(i int, j int) bool {
		return extensions[i].Width > extensions[j].Width
	})

	smallestExtension := extensions[len(extensions)-1]

	widthRemaining := desiredWidth
	for i := 0; i < len(bundle.Gates); i++ {
		currentGate := bundle.Gates[i]
		desiredWidth -= currentGate.Width * float32(currentGate.Qty)
	}

	//widthRemaining := desiredWidth - gate.Width

	index := 0
	for widthRemaining > 0 {
		if index > len(extensions) {
			break
		}

		currentExtension := extensions[index]
		currentExtension.Qty = 1

		if widthRemaining >= currentExtension.Width || (widthRemaining < smallestExtension.Width && widthRemaining > 0) {
			var matchingExtension *types.Extension
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

	bundle.ComputeMetaData()

	return bundle
}
