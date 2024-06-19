package services

import (
	"fmt"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repositories"
)

type CartService struct {
	cartRepo    *repositories.CartRepository
	productRepo *repositories.ProductRepository
}

func (c *CartService) TotalValue(cart models.Cart) (float64, error) {
	value := 0.0
	for _, item := range cart.Items {
		for _, component := range item.Components {
			productPrice, err := c.productRepo.GetPrice(component.ProductID)
			if err != nil {
				return 0, err
			}
			value += (productPrice * float64(component.Qty))
		}
	}
	return value, nil
}

func NewCartService(cartRepo *repositories.CartRepository, productRepo *repositories.ProductRepository) *CartService {
	return &CartService{
		cartRepo,
		productRepo,
	}
}

func (s *CartService) NewCart() (string, error) {
	cart := models.NewCart()
	if _, err := s.cartRepo.SaveCart(cart); err != nil {
		return "", err
	}
	return cart.ID, nil
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

func (s *CartService) AddItem(cartID string, components []models.CartItemComponent) error {
	cartItem := models.NewCartItem(cartID)
	if err := s.cartRepo.SaveCartItem(cartItem); err != nil {
		return err
	}
	if err := s.cartRepo.SaveCartItemComponents(cartItem.ID, components); err != nil {
		return err
	}
	return nil
}

func (s *CartService) RemoveItem(cartID, itemID string) error {
	if err := s.cartRepo.RemoveCartItem(cartID, itemID); err != nil {
		return fmt.Errorf("Failed to remove item. %w", err)
	}
	if err := s.cartRepo.RemoveCartItemComponents(itemID); err != nil {
		return fmt.Errorf("Failed to remove item components. %w", err)
	}
	return nil
}
