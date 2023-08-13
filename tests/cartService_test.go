package tests

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/repositories"
	"github.com/seanomeara96/gates/services"
)

func TestCartService(t *testing.T) {
	db, err := sql.Open("sqlite3", "cart_service_test.db")
	if err != nil {
		t.Error(err)
		return
	}

	_, err = db.Exec("DROP TABLE carts")
	if err != nil {
		t.Error("could not drop tablecarts")
		return
	}
	_, err = db.Exec("DROP TABLE cart_items")
	if err != nil {
		t.Error("could not drop table cart Items")
		return
	}

	cartRepo := repositories.NewCartRepository(db)

	_, err = cartRepo.CreateTables()
	if err != nil {
		t.Error("could not create tables")
		return
	}

	cartService := services.NewCartService(cartRepo)

	userID := 100

	res, err := cartService.NewCart(userID)
	if err != nil {
		t.Error(err)
		return
	}

	lastInsertID, err := res.LastInsertId()
	if err != nil {
		t.Error(err)
		return
	}

	cartID := int(lastInsertID)
	productID := 200

	cartItem, err := cartService.UpdateCartItem(cartID, productID, 1)
	if err != nil {
		t.Error(err)
		return
	}

	if cartItem.CartID != cartID {
		t.Error("cart ID does not match")
	}
	if cartItem.ProductID != productID {
		t.Error("product id is incorrect")
	}
	if cartItem.Quantity != 1 {
		t.Error("product qty is incorrect")
	}

	cart, cartItems, err := cartService.GetCart(userID)
	if err != nil {
		t.Error("could not get cart and cartItems")
		return
	}

	if cart.UserID != userID {
		t.Error("incorrect user ID")
	}

	for _, item := range cartItems {
		if item.CartID != cart.ID {
			t.Error("cart item does not belong to cart")
		}
	}

	if cartItems[0].ProductID != productID {
		t.Error("incorrect product ID")
	}

	cartItem, err = cartService.UpdateCartItem(cart.ID, productID, 2)
	if err != nil {
		t.Error("could not increment cart item")
		return
	}

	if cartItem.Quantity != 3 {
		t.Errorf("expected quantity to be 3 got %d instead", cartItem.Quantity)
	}

	cartItem, err = cartService.UpdateCartItem(cart.ID, productID, -6)
	if err != nil {
		t.Error("could not decrement cart Item")
		return
	}

	if cartItem.Quantity != 0 {
		t.Errorf("expected qty to be 0 got %d instead", cartItem.Quantity)
	}

	cartItem, err = cartService.UpdateCartItem(cartID, productID, 6)
	if err != nil {
		t.Error("could not update cartItem")
		return
	}

	if cartItem.Quantity != 6 {
		t.Errorf("expected 6 got %d", cartItem.Quantity)
		return
	}

	err = cartService.RemoveCartItem(cartID, productID)
	if err != nil {
		t.Error("could not remove cart item")
		return
	}

	_, err = cartService.UpdateCartItem(cartID, 201, 1)
	if err != nil {
		t.Error("could not add cart items")
		return
	}

	_, err = cartService.UpdateCartItem(cartID, 202, 2)

	if err != nil {
		t.Error("could not add cart items")
		return
	}

	err = cartService.RemoveAllCartItems(cartID)
	if err != nil {
		t.Error("could not remove all cart items")
	}

	_, cartItems, err = cartService.GetCart(userID)
	if err != nil {
		t.Error("could not get cart for user")
	}
	for _, item := range cartItems {
		if item.Quantity != 0 {
			t.Errorf("expected a qty of 0. got %d instead", item.Quantity)
		}
	}
}
