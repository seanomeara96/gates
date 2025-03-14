package models

import (
	"strconv"
)

type Bundle struct {
	Product
	Components []Product `json:"components"`
}

// we can add in error handling retrospectively,
// for now, lets assume we're not doing anything stupid
func (b *Bundle) ComputeMetaData() {
	b.setType()
	b.setName()
	b.setPrice()
	b.setImg()
	b.setColor()
	b.setQty()
	b.setWidth()
	b.setTolerance()
}

func (b *Bundle) ToProduct() Product {
	b.ComputeMetaData()
	return Product{
		Type:      "bundle",
		Name:      b.Name,
		Width:     b.Width,
		Price:     b.Price,
		Img:       b.Img,
		Color:     b.Color,
		Tolerance: b.Tolerance,
		Qty:       b.Qty,
	}
}

/*
	for all of the below, we assume that product at index 0 is the gate and that there is
	only one gate per bundle
*/

func (b *Bundle) setType() {
	b.Type = "bundle"
}

func (b *Bundle) setQty() {
	if b.Qty < 1 {
		b.Qty = 1
	}
}

func (b *Bundle) setTolerance() {
	if b.Tolerance == 0 {
		b.Tolerance = b.Components[0].Tolerance
	}
}

func (b *Bundle) setPrice() {
	if b.Price == 0 {
		for i := 0; i < len(b.Components); i++ {
			b.Price += (b.Components[i].Price * float32(b.Components[i].Qty))
		}
	}
}
func (b *Bundle) setImg() {
	if b.Img == "" {
		b.Img = b.Components[0].Img
	}
}
func (b *Bundle) setColor() {
	if b.Color == "" {
		b.Color = b.Components[0].Color
	}
}
func (b *Bundle) setName() {
	if b.Name == "" {
		gate := b.Components[0]
		b.Name = gate.Name
		extensionCount := 0
		for i := 1; i < len(b.Components); i++ {
			extensionCount += b.Components[i].Qty
		}
		if extensionCount == 1 {
			b.Name += " and 1 extension."
		} else if extensionCount > 1 {
			b.Name += " and " + strconv.Itoa(extensionCount) + " extensions."
		}

	}
}

func (b *Bundle) setWidth() {
	if b.Width == 0 {
		b.Width = b.Components[0].Width * float32(b.Components[0].Qty)
		for i := 1; i < len(b.Components); i++ {
			b.Width += (b.Components[i].Width * float32(b.Components[i].Qty))
		}
	}
}
