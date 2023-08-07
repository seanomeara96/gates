package services

import "github.com/seanomeara96/gates/repositories"

type CartService struct {
	cartRepo *repositories.CartRepository
}

func NewCartService(repo *repositories.CartRepository) *CartService {
	return &CartService{
		repo,
	}
}
