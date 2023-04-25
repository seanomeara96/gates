package types

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
	MaxWidth   float32     `json:"max_width"`
}

func (b *Bundle) ComputeMetaData() {
	for i := 0; i < len(b.Gates); i++ {
		gate := b.Gates[i]
		b.MaxWidth += gate.Width * float32(gate.Qty)
		b.Price += gate.Price * float32(gate.Qty)
	}
	for i := 0; i < len(b.Extensions); i++ {
		extension := b.Extensions[i]
		b.MaxWidth += extension.Width * float32(extension.Qty)
		b.Price += extension.Price * float32(extension.Qty)
	}
	if b.Qty < 1 {
		b.Qty = 1
	}
}

type Bundles []Bundle
