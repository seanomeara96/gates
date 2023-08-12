package services

import (
	"database/sql"

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

func (s *CartService) NewCart(userID int) (sql.Result, error) {
	return s.cartRepo.SaveCart(models.NewCart(userID))
}

func (s *CartService) IncrementCartItem(cartID, productID, qty int) (*models.CartItem, error) {
	cartItemToAdd, err := s.cartRepo.GetCartItemByProductID(cartID, productID)
	if err != nil {
		// doing it this way is going to cause problems because errors are thrown for reasons other than row not found
		newItem := models.NewCartItem(cartID, productID, qty)
		res, err := s.cartRepo.SaveCartItem(newItem)
		if err != nil {
			return nil, err
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
		newItem.ID = int(lastID)
		return &newItem, err
	}
	cartItemToAdd.Quantity += qty
	_, err = s.cartRepo.UpdateCartItem(*cartItemToAdd)
	return cartItemToAdd, err
}

func (s *CartService) DecrementCartItem(cartID, productID, qty int) (*models.CartItem, error) {
	cartItemToAdd, err := s.cartRepo.GetCartItemByProductID(cartID, productID)
	if err != nil {
		return nil, err
	}
	cartItemToAdd.Quantity -= qty
	if cartItemToAdd.Quantity < 0 {
		cartItemToAdd.Quantity = 0
	}
	_, err = s.cartRepo.UpdateCartItem(*cartItemToAdd)
	return cartItemToAdd, err
}

func (s *CartService) GetCart(userID int) (*models.Cart, []*models.CartItem, error) {
	cart, err := s.cartRepo.GetCartByUserID(userID)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.cartRepo.GetCartItemsByCartID(cart.ID)
	if err != nil {
		return nil, nil, err
	}
	return cart, items, nil
}
