package models

type Gate struct {
	Product
	Tolerance float32 `json:"tolerance"`
}

type Gates []Gate
