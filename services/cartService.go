package services

import (
	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repositories"
)

type CartService struct {
	cartRepo *repositories.CartRepository
}

func NewCartService(repo *repositories.CartRepository) *CartService {
	return &CartService{
		repo,
	}
}

func (s *CartService) NewCart(userID int) (models.Cart, error) {
	return s.cartRepo.SaveCart(models.NewCart(userID))
}

func (s *CartService) AddCartItem(cartID, productID, qty int) (models.CartItem, error) {
	return s.cartRepo.SaveCartItem(models.NewCartItem(cartID, productID, qty))
}

func (s *CartService) UpdateCartItem(cartID, productID, qty int) (models.CartItem, error) {
	return s.cartRepo.SaveCartItem(models.NewCartItem(cartID, productID, qty))
}

func (s *CartService) GetCart(userID int) (*models.Cart, []*models.CartItem, error) {
	cart, err := s.cartRepo.GetCartByUserId(userID)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.cartRepo.GetCartItemsByCartId(cart.ID)
	if err != nil {
		return nil, nil, err
	}
	return cart, items, nil
}
