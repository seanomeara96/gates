package main

import "strconv"

type Product struct {
	Id    int     `json:"id"`
	Name  string  `json:"name"`
	Width float32 `json:"width"`
	Price float32 `json:"price"`
	Img   string  `json:"img"`
	Color string  `json:"color"`
	Qty   int     `json:"qty"`
}

type Gate struct {
	Product
	Tolerance float32 `json:"tolerance"`
}

type Gates []Gate

type Extension struct {
	Product
}

type Extensions []Extension

type Bundle struct {
	Product
	Gates      []Gate      `json:"gates"`
	Extensions []Extension `json:"extensions"`
	Tolerance  float32     `json:"tolerance"`
}

type Bundles []Bundle

// we can add in error handling retrospectively,
// for now, lets assume we're not doing anything stupid
func (b *Bundle) ComputeMetaData() {
	b.setName()
	b.setPrice()
	b.setImg()
	b.setColor()
	b.setQty()
	b.setWidth()
	b.setTolerance()
}

func (b *Bundle) setQty() {
	if b.Qty < 1 {
		b.Qty = 1
	}
}

func (b *Bundle) setTolerance() {
	b.Tolerance = b.Gates[0].Tolerance
}

func (b *Bundle) setPrice() {
	for i := 0; i < len(b.Gates); i++ {
		b.Price += (b.Gates[i].Price * float32(b.Gates[i].Qty))
	}
	for ii := 0; ii < len(b.Extensions); ii++ {
		b.Price += (b.Extensions[ii].Price * float32(b.Extensions[ii].Qty))
	}
}
func (b *Bundle) setImg() {
	b.Img = b.Gates[0].Img
}
func (b *Bundle) setColor() {
	b.Color = b.Gates[0].Color
}
func (b *Bundle) setName() {
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

func (b *Bundle) setWidth() {
	b.Width = b.Gates[0].Width
	for i := 0; i < len(b.Extensions); i++ {
		b.Width += (b.Extensions[i].Width * float32(b.Extensions[i].Qty))
	}
}
