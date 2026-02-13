package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/seanomeara96/gates/models"
)

func (h *Handler) FetchOrderDetailsFromStripe(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	orderID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("parse order id from path: %w", err)
	}

	order, err := h.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return fmt.Errorf("get order by id %d: %w", orderID, err)
	}

	fmt.Printf("[DEV] order details to fetched from stripe. not yet implemented %d", order.ID)
	// use order

	return errors.New("fetch order details from stripe: not yet implemented")

}
