package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/seanomeara96/gates/models"
)

func (h *Handler) UpdateOrder(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return err
	}
	order, err := h.orderRepo.GetOrderByID(id)
	if err != nil {
		return err
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		return err
	}

	// Update order fields from form data
	if status := r.FormValue("status"); status != "" {
		order.Status = models.OrderStatus(status)
	}

	if customerName := r.FormValue("customer_name"); customerName != "" {
		order.CustomerName.String = customerName
		order.CustomerName.Valid = true
	}

	if customerEmail := r.FormValue("customer_email"); customerEmail != "" {
		order.CustomerEmail.String = customerEmail
		order.CustomerEmail.Valid = true
	}

	if customerPhone := r.FormValue("customer_phone"); customerPhone != "" {
		order.CustomerPhone.String = customerPhone
		order.CustomerPhone.Valid = true
	}

	if shippingAddress := r.FormValue("shipping_address"); shippingAddress != "" {
		order.ShippingAddress.String = shippingAddress
		order.ShippingAddress.Valid = true
	}

	if billingAddress := r.FormValue("billing_address"); billingAddress != "" {
		order.BillingAddress.String = billingAddress
		order.BillingAddress.Valid = true
	}

	if paymentMethod := r.FormValue("payment_method"); paymentMethod != "" {
		order.PaymentMethod.String = paymentMethod
		order.PaymentMethod.Valid = true
	}

	if stripeRef := r.FormValue("stripe_ref"); stripeRef != "" {
		order.StripeRef.String = stripeRef
		order.StripeRef.Valid = true
	}

	if sessionID := r.FormValue("session_id"); sessionID != "" {
		order.SessionID.String = sessionID
		order.SessionID.Valid = true
	}

	return h.orderRepo.UpdateOrder(order)
}

func (h *Handler) UpdateOrderStatus(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return err
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		return err
	}

	// Get status from form
	status := r.FormValue("status")
	if status == "" {
		return fmt.Errorf("status is required")
	}

	// Update order status
	return h.orderRepo.UpdateStatus(id, models.OrderStatus(status))
}
