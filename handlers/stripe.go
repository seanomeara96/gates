package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/seanomeara96/gates/models"
)

func (h *Handler) FetchOrderDetailsFromStripe(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	orderID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return err
	}

	order, err := h.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	fmt.Printf("[DEV] order details to fetched from stripe. not yet implemented %d", order.ID)
	// use order

	return errors.New("not yet implemented")

}
