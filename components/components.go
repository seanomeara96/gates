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
