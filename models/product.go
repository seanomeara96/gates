package models

type Product struct {
	Id    int     `json:"id"`
	Name  string  `json:"name"`
	Width float32 `json:"width"`
	Price float32 `json:"price"`
	Img   string  `json:"img"`
	Color string  `json:"color"`
	Qty   int     `json:"qty"`
}
