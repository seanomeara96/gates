package handlers

import (
	"fmt"
	"net/http"

	"github.com/seanomeara96/gates/models"
)

type MiddlewareFunc func(next CustomHandleFunc) CustomHandleFunc

func (h *Handler) GetCartFromRequest(next CustomHandleFunc) CustomHandleFunc {
	return func(_ models.Cart, w http.ResponseWriter, r *http.Request) error {
		// returns new session if does not exist
		session, err := getCartSession(r, h.cookieStore)
		if err != nil {
			return fmt.Errorf("cart middleware: failed to get cart session from cookie store: %w", err)
		}

		cartID, cartIDExists, err := getCartID(session)
		if err != nil {
			return fmt.Errorf("cart middleware: failed to get cart ID from session: %w", err)
		}

		if cartIDExists {
			cart, cartExists, err := h.cartRepo.GetCartByID(cartID)
			if err != nil {
				return fmt.Errorf("cart middleware: failed to get cart by ID %q: %w", cartID, err)
			}
			if cartExists {
				return next(cart, w, r)
			}
		}

		cart, err := h.newCart()
		if err != nil {
			return fmt.Errorf("cart middleware: failed to create a new cart: %w", err)
		}
		if err := attachNewCartToSession(cart, session, w, r); err != nil {
			return fmt.Errorf("cart middleware: failed to attach new cart to session: %w", err)
		}

		return next(cart, w, r)
	}
}

func (h *Handler) MustBeAdmin(next CustomHandleFunc) CustomHandleFunc {
	return func(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
		accessToken, refreshToken, err := h.auth.GetTokensFromRequest(r)
		if err != nil {
			return fmt.Errorf("MustBeAdmin middleware: failed to get tokens from request: %w", err)
		}
		claims, err := h.auth.ValidateToken(accessToken)
		if err != nil {
			accessToken, refreshToken, err = h.auth.Refresh(r.Context(), refreshToken)
			if err != nil {
				return fmt.Errorf("MustBeAdmin middleware: failed to refresh access token: %w", err)
			}
			claims, err = h.auth.ValidateToken(accessToken)
			if err != nil {
				return fmt.Errorf("MustBeAdmin middleware: failed to validate refreshed access token: %w", err)
			}
		}
		/*
			we have an issue here
			claims should include the public user id on the auth user struct
			but instead contains the int id
			this needs to be fixed before we can move forward
		*/
		if claims.UserID != h.cfg.AdminUserID {
			return fmt.Errorf("MustBeAdmin middleware: user is not admin (userID=%v, adminUserID=%v)", claims.UserID, h.cfg.AdminUserID)
		}
		return next(cart, w, r)
	}
}
