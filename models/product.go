package models

type Product struct {
	Id        int     `json:"id"`
	Type      string  `json:"type"`
	Name      string  `json:"name"`
	Width     float32 `json:"width"`
	Price     float32 `json:"price"`
	Img       string  `json:"img"`
	Color     string  `json:"color"`
	Tolerance float32 `json:"tolerance"`
	Qty       int     `json:"qty"`
}
