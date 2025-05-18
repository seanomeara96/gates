package models

import (
	"database/sql"
	"fmt"
	"time"
)

// Define a custom type for OrderStatus for better type safety
type OrderStatus string

// Define constants for common order statuses
const (
	// Initial statuses
	OrderStatusPendingPayment  OrderStatus = "pending_payment"  // Order placed, awaiting payment confirmation
	OrderStatusAwaitingPayment OrderStatus = "awaiting_payment" // Customer completed checkout, payment pending
	OrderStatusDraft           OrderStatus = "draft"            // Order created manually, not yet submitted

	// In-progress statuses
	OrderStatusProcessing          OrderStatus = "processing"           // Payment received, order being prepared for fulfillment
	OrderStatusOnHold              OrderStatus = "on_hold"              // Order temporarily paused (e.g., for review, stock issue)
	OrderStatusAwaitingFulfillment OrderStatus = "awaiting_fulfillment" // Paid, waiting for picking/packing
	OrderStatusAwaitingShipment    OrderStatus = "awaiting_shipment"    // Packed, waiting for carrier pickup
	OrderStatusPartiallyShipped    OrderStatus = "partially_shipped"    // Some items shipped
	OrderStatusShipped             OrderStatus = "shipped"              // All items shipped
	OrderStatusOutForDelivery      OrderStatus = "out_for_delivery"     // With local carrier for delivery
	OrderStatusAwaitingPickup      OrderStatus = "awaiting_pickup"      // Ready for customer pickup

	// Completed/Terminal statuses
	OrderStatusCompleted       OrderStatus = "completed"        // Order successfully fulfilled and closed
	OrderStatusDelivered       OrderStatus = "delivered"        // Package successfully delivered to customer
	OrderStatusPickedUp        OrderStatus = "picked_up"        // Customer has picked up the order
	OrderStatusCanceled        OrderStatus = "canceled"         // Order canceled before completion
	OrderStatusFailed          OrderStatus = "failed"           // Payment failed
	OrderStatusRefunded        OrderStatus = "refunded"         // Full refund issued
	OrderStatusPartialRefunded OrderStatus = "partial_refunded" // Partial refund issued
	OrderStatusClosed          OrderStatus = "closed"           // Order finalized and closed (can be similar to completed)

	// Exception statuses
	OrderStatusFraud      OrderStatus = "fraud"      // Order flagged as potentially fraudulent
	OrderStatusChargeback OrderStatus = "chargeback" // Customer initiated a chargeback
	OrderStatusError      OrderStatus = "error"      // An error occurred during processing
)

// validOrderStatuses is a map to quickly check if a status is valid.
// Using struct{} as the value type is memory efficient as it takes zero bytes.
var validOrderStatuses = map[OrderStatus]struct{}{
	OrderStatusPendingPayment:      {},
	OrderStatusAwaitingPayment:     {},
	OrderStatusDraft:               {},
	OrderStatusProcessing:          {},
	OrderStatusOnHold:              {},
	OrderStatusAwaitingFulfillment: {},
	OrderStatusAwaitingShipment:    {},
	OrderStatusPartiallyShipped:    {},
	OrderStatusShipped:             {},
	OrderStatusOutForDelivery:      {},
	OrderStatusAwaitingPickup:      {},
	OrderStatusCompleted:           {},
	OrderStatusDelivered:           {},
	OrderStatusPickedUp:            {},
	OrderStatusCanceled:            {},
	OrderStatusFailed:              {},
	OrderStatusRefunded:            {},
	OrderStatusPartialRefunded:     {},
	OrderStatusClosed:              {},
	OrderStatusFraud:               {},
	OrderStatusChargeback:          {},
	OrderStatusError:               {},
}

// IsValid checks if the given OrderStatus is one of the predefined valid statuses.
func (s OrderStatus) IsValid() bool {
	_, ok := validOrderStatuses[s]
	return ok
}

// Validate checks if the given OrderStatus is valid and returns an error if not.
func (s OrderStatus) Validate() error {
	if !s.IsValid() {
		return fmt.Errorf("invalid order status: %q", s)
	}
	return nil
}

type Order struct {
	ID              int
	Status          string
	CartID          string
	SessionID       string
	CustomerName    sql.NullString
	CustomerPhone   string
	ShippingAddress string
	BillingAddress  string
	PaymentMethod   string
	CreatedAt       time.Time
}
