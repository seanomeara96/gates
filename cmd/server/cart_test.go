package main

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/seanomeara96/gates/models"
)

var db *sql.DB
var cart models.Cart

func TestMain(m *testing.M) {
	db = SqliteOpen("../../main.db")
	defer db.Close()

	cart = models.NewCart()

	code := m.Run()

	os.Exit(code)
}

func TestSaveCart(t *testing.T) {
	_, err := SaveCart(db, cart)
	if err != nil {
		t.Error(err)
	}
}

func TestGetCart(t *testing.T) {
	resCart, err := GetCartByID(db, cart.ID)
	if err != nil {
		t.Error(err)
		return
	}
	if cart.ID != resCart.ID {
		t.Errorf("expected cart with id %s, got %s instead", cart.ID, resCart.ID)
	}
}

func TestAddItemToCart(t *testing.T) {
	item := models.NewCartItem(cart.ID, []models.CartItemComponent{
		models.CartItemComponent{ProductID: 1, Qty: 1, CreatedAt: time.Now()},
	})

	err := AddItemToCart(db, cart.ID, item)
	if err != nil {
		t.Error(err)
		return
	}

	resCart, err := GetCartByID(db, cart.ID)
	if err != nil {
		t.Error(err)
		return
	}

	if len(resCart.Items) < 1 {
		t.Errorf("expected an item in the cart instead got %d", len(resCart.Items))
		return
	}

	if resCart.Items[0].ID != item.ID {
		t.Error("expected item and db item to have same ID")
		return
	}

	cart = *resCart

	err = AddItemToCart(db, cart.ID, item)
	if err != nil {
		t.Error(err)
		return
	}

	resCart, err = GetCartByID(db, cart.ID)
	if err != nil {
		t.Error(err)
		return
	}

	if len(resCart.Items) > 1 {
		t.Error("there should only be one item in the cart")
	}

	if resCart.Items[0].Qty < 2 {
		t.Error("the item qty should have been incremented")
	}

	cart = *resCart

	bundle := models.Bundle{
		Gates: []models.Product{
			models.Product{
				Id:  1,
				Qty: 1,
			},
		},
		Extensions: []models.Product{
			models.Product{
				Id:  3,
				Qty: 1,
			},
		},
	}
	bundleComponents := bundle.Components()

	bundleAsItem := models.NewCartItem(cart.ID, bundleComponents)

	err = AddItemToCart(db, cart.ID, bundleAsItem)
	if err != nil {
		t.Error(err)
		return
	}

	resCart, err = GetCartByID(db, cart.ID)
	if err != nil {
		t.Error(err)
		return
	}

	if len(resCart.Items) > 2 {
		t.Error("there should only be two items in the cart")
	}

	foundBundle := false
	for _, item := range resCart.Items {
		if len(item.Components) > 1 {
			foundBundle = true
		}
		if len(item.Components) < 1 {
			t.Error("an item has no components")
			return
		}
	}

	if !foundBundle {
		t.Error("there should deffo be a bundle")
	}

	cart = *resCart

}
