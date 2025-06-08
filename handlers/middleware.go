package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/seanomeara96/gates/models"
)

type MiddlewareFunc func(next CustomHandleFunc) CustomHandleFunc

func (h *Handler) GetCartFromRequest(next CustomHandleFunc) CustomHandleFunc {
	return func(_ *models.Cart, w http.ResponseWriter, r *http.Request) error {
		// returns new session if does not exist
		session, err := getCartSession(r, h.cookieStore)
		if err != nil {
			return fmt.Errorf("Failed to get cart session from cookie store in cart middleware: %w", err)
		}

		cartID, cartIDExists, err := getCartID(session)
		if err != nil {
			return fmt.Errorf("Failed to get cart ID from session in cart middleware: %w", err)
		}

		if cartIDExists {
			cart, cartExists, err := h.cartRepo.GetCartByID(cartID)
			if err != nil {
				return fmt.Errorf("failed to get cart by ID %w", err)
			}
			if cartExists {
				return next(cart, w, r)
			}
		}

		cart, err := h.newCart()
		if err != nil {
			return fmt.Errorf("Failure to create a new cart in the cart middleware: %w", err)
		}
		if err := attachNewCartToSession(cart, session, w, r); err != nil {
			return fmt.Errorf("Failed to attach new cart to session in cart middlware: %w;", err)
		}

		return next(cart, w, r)
	}
}

func (h *Handler) MustBeAdmin(next CustomHandleFunc) CustomHandleFunc {
	return func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		accessToken, refreshToken, err := h.auth.GetTokensFromRequest(r)
		if err != nil {
			return fmt.Errorf(". Cant get tokens from request in MustBeAdmin middleware func: %w", err)
		}
		claims, err := h.auth.ValidateToken(accessToken)
		if err != nil {
			accessToken, refreshToken, err = h.auth.Refresh(r.Context(), refreshToken)
			if err != nil {
				return err
			}
		}
		/*
			we have an issue here
			claims should include the public user id on the auth user struct
			but instead contains the int id
			this needs to be fixed before we can move forward
		*/
		if claims.UserID != h.cfg.AdminUserID {
			return errors.New("user is not admin")
		}
		return next(cart, w, r)
	}
}
