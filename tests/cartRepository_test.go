package tests

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repositories"
)

func Test(t *testing.T) {
	db, err := sql.Open("sqlite3", "carts_test.db")
	if err != nil {
		t.Error("could not connect to database")
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

	repo := repositories.NewCartRepository(db)
	_, err = repo.CreateTables()
	if err != nil {
		t.Error("could not create tables")
		return
	}

	// Simulate creating a few carts with cart items
	carts := []models.Cart{
		{ID: 1, UserID: 101, CreatedAt: time.Now(), LastUpdatedAt: time.Now()},
		{ID: 2, UserID: 102, CreatedAt: time.Now(), LastUpdatedAt: time.Now()},
	}

	cartItems := []models.CartItem{
		{ID: 1, CartID: 1, ProductID: 201, Quantity: 2, CreatedAt: time.Now()},
		{ID: 2, CartID: 1, ProductID: 202, Quantity: 3, CreatedAt: time.Now()},
		{ID: 3, CartID: 2, ProductID: 203, Quantity: 1, CreatedAt: time.Now()},
	}

	for _, cart := range carts {
		_, err := repo.SaveCart(cart)
		if err != nil {
			t.Error("could not save cart")
			return
		}
		for _, cartItem := range cartItems {
			if cartItem.CartID == cart.ID {
				_, err := repo.SaveCartItem(cartItem)
				if err != nil {
					t.Error("could not save cart items")
					return
				}
			}
		}
	}

	cart, err := repo.GetCartByUserID(101)
	if err != nil {
		t.Error("could not get cart by user id")
		fmt.Println(err)
		return
	}

	if cart.UserID != 101 || cart.ID != 1 {
		t.Error("incorrect cart returned")
		return
	}

	cartsItems, err := repo.GetCartItemsByCartID(cart.ID)
	if err != nil {
		t.Error("could not get cart's items")
		return
	}
	for _, cartItem := range cartsItems {
		if cartItem.CartID != cart.ID {
			t.Error("this cartItem does not belong in this cart")
		}
	}

	matchingCart, err := repo.GetCartItemByID(cartItems[0].ID)
	if err != nil {
		t.Error("error occured tring to find matching cart")
		return
	}

	if cartItems[0].ID != matchingCart.ID {
		t.Error("cartItem Ids dont match")
		return
	}

	product, err := repo.GetCartItemByProductID(cartItems[0].CartID, cartItems[0].ProductID)
	if err != nil {
		t.Error("could not fetch cart item by product id")
		return
	}

	if product.ID != cartItems[0].ID || product.ProductID != cartItems[0].ProductID {
		t.Error("incorrect item fetched")
		return
	}

}
