package models

import (
	"strconv"
)

type Bundle struct {
	Product
	Gates      []Product `json:"gates"`
	Extensions []Product `json:"extensions"`
}

type Bundles []Bundle

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
		b.Tolerance = b.Gates[0].Tolerance
	}
}

func (b *Bundle) setPrice() {
	if b.Price == 0 {
		for i := 0; i < len(b.Gates); i++ {
			b.Price += (b.Gates[i].Price * float32(b.Gates[i].Qty))
		}
		for ii := 0; ii < len(b.Extensions); ii++ {
			b.Price += (b.Extensions[ii].Price * float32(b.Extensions[ii].Qty))
		}
	}
}
func (b *Bundle) setImg() {
	if b.Img == "" {
		b.Img = b.Gates[0].Img
	}
}
func (b *Bundle) setColor() {
	if b.Color == "" {
		b.Color = b.Gates[0].Color
	}
}
func (b *Bundle) setName() {
	if b.Name == "" {
		gate := b.Gates[0]
		b.Name = gate.Name
		extensionCount := 0
		for i := 0; i < len(b.Extensions); i++ {
			extensionCount += b.Extensions[i].Qty
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
		b.Width = b.Gates[0].Width * float32(b.Gates[0].Qty)
		for i := 0; i < len(b.Extensions); i++ {
			b.Width += (b.Extensions[i].Width * float32(b.Extensions[i].Qty))
		}
	}
}
