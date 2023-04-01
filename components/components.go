package components

type Gate struct {
	Id        int     `json:"id"`
	Name      string  `json:"name"`
	Width     float32 `json:"width"`
	Price     float32 `json:"price"`
	Img       string  `json:"img"`
	Tolerance float32 `json:"tolerance"`
	Color     string  `json:"color"`
	Qty       int     `json:"qty"`
}

type Gates []Gate

type Extension struct {
	Id    int     `json:"id"`
	Name  string  `json:"name"`
	Width float32 `json:"width"`
	Price float32 `json:"price"`
	Img   string  `json:"img"`
	Color string  `json:"color"`
	Qty   int     `json:"qty"`
}

type Extensions []Extension

type Bundle struct {
	Gate       Gate       `json:"gate"`
	Extensions Extensions `json:"extensions"`
	Price      float32    `json:"price"`
	MaxLength  float32    `json:"max_length"`
	Qty        int        `json:"qty"`
}

type Bundles []Bundle

type CachedBundle struct {
	Id    int
	Name  string
	Size  float32
	Price float32
	Color string
}

type CachedBundles []CachedBundle

func (bundle *Bundle) ComputeMetaData() {
	bundle.MaxLength = bundle.Gate.Width
	bundle.Price = bundle.Gate.Price
	for _, extension := range bundle.Extensions {
		bundle.MaxLength += extension.Width * float32(extension.Qty)
		bundle.Price += extension.Price * float32(extension.Qty)
	}

	// need to test if the width of the bundle minus the threshold actuall accomodates the requested size
	// if it doesnt set the qty of the bundle to 0
	bundle.Qty = 1
}
