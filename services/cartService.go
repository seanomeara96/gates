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

func (s *CartService) NewCart() (string, error) {
	cart := models.NewCart()
	if _, err := s.cartRepo.SaveCart(cart); err != nil {
		return "", err
	}
	return cart.ID, nil
}

func (s *CartService) UpdateCartItem(cartID string, productID, qty int) (*models.CartItem, error) {
	cartItemToAdd, err := s.cartRepo.GetCartItemByProductID(cartID, productID)
	if err != nil {
		// doing it this way is going to cause problems because errors are thrown for reasons other than row not found
		if qty < 0 {
			qty = 0
		}
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
	if cartItemToAdd.Quantity < 0 {
		cartItemToAdd.Quantity = 0
	}
	_, err = s.cartRepo.UpdateCartItem(*cartItemToAdd)
	return cartItemToAdd, err
}

func (s *CartService) GetCart(userID int) (*models.Cart, []*models.CartItem, error) {
	//maybe if there is not cartby that user id we should auto call new cart
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

func (s *CartService) RemoveCartItem(cartID string, productID int) error {
	cartItem, err := s.cartRepo.GetCartItemByProductID(cartID, productID)
	if err != nil {
		return err
	}
	_, err = s.UpdateCartItem(cartID, productID, -cartItem.Quantity)
	if err != nil {
		return err
	}
	return nil
}

func (s *CartService) RemoveAllCartItems(cartID string) error {
	cartItems, err := s.cartRepo.GetCartItemsByCartID(cartID)
	if err != nil {
		return err
	}
	for _, item := range cartItems {
		err = s.RemoveCartItem(cartID, item.ProductID)
		if err != nil {
			return err
		}
	}
	return nil
}
