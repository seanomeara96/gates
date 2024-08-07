package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

func TestCartPrepare(t *testing.T) {
	cartData := `{
  "id": "536255d4-d843-4a39-ac69-b63c9402fcb6",
  "created_at": "2024-07-22T20:40:41.455095173+01:00",
  "last_updated_at": "2024-08-07T20:30:06.724698339+01:00",
  "items": [
    {
      "id": "1-1",
      "cart_id": "536255d4-d843-4a39-ac69-b63c9402fcb6",
      "name": "",
      "sale_price": 0,
      "components": [
        {
          "cart_item_id": "1-1",
          "cart_id": "536255d4-d843-4a39-ac69-b63c9402fcb6",
          "product_id": 1,
          "qty": 1,
          "name": "",
          "created_at": "2024-08-07T20:29:46.457331985+01:00"
        }
      ],
      "qty": 1,
      "created_at": "2024-08-07T20:29:46.457346354+01:00"
    },
    {
      "id": "2-1",
      "cart_id": "536255d4-d843-4a39-ac69-b63c9402fcb6",
      "name": "",
      "sale_price": 0,
      "components": [
        {
          "cart_item_id": "2-1",
          "cart_id": "536255d4-d843-4a39-ac69-b63c9402fcb6",
          "product_id": 2,
          "qty": 1,
          "name": "",
          "created_at": "2024-08-07T20:29:47.420480358+01:00"
        }
      ],
      "qty": 1,
      "created_at": "2024-08-07T20:29:47.420494565+01:00"
    },
    {
      "id": "3-1",
      "cart_id": "536255d4-d843-4a39-ac69-b63c9402fcb6",
      "name": "",
      "sale_price": 0,
      "components": [
        {
          "cart_item_id": "3-1",
          "cart_id": "536255d4-d843-4a39-ac69-b63c9402fcb6",
          "product_id": 3,
          "qty": 1,
          "name": "",
          "created_at": "2024-08-07T20:29:48.435215408+01:00"
        }
      ],
      "qty": 4,
      "created_at": "2024-08-07T20:29:48.4352302+01:00"
    },
    {
      "id": "4-1",
      "cart_id": "536255d4-d843-4a39-ac69-b63c9402fcb6",
      "name": "",
      "sale_price": 0,
      "components": [
        {
          "cart_item_id": "4-1",
          "cart_id": "536255d4-d843-4a39-ac69-b63c9402fcb6",
          "product_id": 4,
          "qty": 1,
          "name": "",
          "created_at": "2024-08-07T20:29:49.623132319+01:00"
        }
      ],
      "qty": 1,
      "created_at": "2024-08-07T20:29:49.623145941+01:00"
    },
    {
      "id": "1-1_4-1_5-2",
      "cart_id": "536255d4-d843-4a39-ac69-b63c9402fcb6",
      "name": "",
      "sale_price": 0,
      "components": [
        {
          "cart_item_id": "1-1_4-1_5-2",
          "cart_id": "536255d4-d843-4a39-ac69-b63c9402fcb6",
          "product_id": 1,
          "qty": 1,
          "name": "",
          "created_at": "2024-08-07T20:30:06.676876438+01:00"
        },
        {
          "cart_item_id": "1-1_4-1_5-2",
          "cart_id": "536255d4-d843-4a39-ac69-b63c9402fcb6",
          "product_id": 4,
          "qty": 1,
          "name": "",
          "created_at": "2024-08-07T20:30:06.676889805+01:00"
        },
        {
          "cart_item_id": "1-1_4-1_5-2",
          "cart_id": "536255d4-d843-4a39-ac69-b63c9402fcb6",
          "product_id": 5,
          "qty": 2,
          "name": "",
          "created_at": "2024-08-07T20:30:06.676892541+01:00"
        }
      ],
      "qty": 13,
      "created_at": "2024-08-07T20:30:06.676896868+01:00"
    }
  ],
  "total_value": 0
}`

	var cart models.Cart
	if err := json.Unmarshal([]byte(cartData), &cart); err != nil {
		t.Error(err)
	}

	if err := prepareCart(db, &cart); err != nil {
		t.Error(err)
	}

	for _, i := range cart.Items {
		fmt.Println(i.Qty, i.SalePrice, float64(i.Qty)*i.SalePrice)
	}

	t.Error(cart.TotalValue)
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
		{ProductID: 1, Qty: 1, CreatedAt: time.Now()},
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
			{
				Id:  1,
				Qty: 1,
			},
		},
		Extensions: []models.Product{
			{
				Id:  3,
				Qty: 1,
			},
		},
	}

	bundleAsItem := models.NewCartItem(cart.ID, bundle.CartItemComponents())

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
			fmt.Println(item)
			t.Error("an item has no components")
			return
		}
	}

	if !foundBundle {
		t.Error("there should deffo be a bundle")
	}

	cart = *resCart

}
