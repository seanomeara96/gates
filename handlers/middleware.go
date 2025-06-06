package handlers

import (
	"fmt"
	"net/http"

	"github.com/seanomeara96/gates/models"
)

type MiddlewareFunc func(next CustomHandleFunc) CustomHandleFunc

func (h *Handler) GetCartFromRequest(next CustomHandleFunc) CustomHandleFunc {
	return func(_ *models.Cart, w http.ResponseWriter, r *http.Request) error {
		session, err := getCartSession(r, h.cookieStore)
		if err != nil {
			return err
		}

		cartID, err := getCartID(session)
		if err != nil {
			return err
		}

		if cartID == nil {
			cart, err := h.newCart()
			if err != nil {
				return err
			}
			if err := attachNewCartToSession(cart, session, w, r); err != nil {
				return err
			}
			return nil
		}

		if valid := validateCartID(cartID); !valid {
			return fmt.Errorf("cart id is invalid")
		}

		exists, err := h.cartRepo.CartExists(cartID.(string))
		if err != nil {
			return err
		}
		if !exists {
			cart, err := h.newCart()
			if err != nil {
				return err
			}
			if err := attachNewCartToSession(cart, session, w, r); err != nil {
				return err
			}
			return next(cart, w, r)
		}

		cart, err := h.cartRepo.GetCartByID(cartID.(string))
		if err != nil {
			return err
		}

		return next(cart, w, r)
	}
}

func (h *Handler) MustBeAdmin(next CustomHandleFunc) CustomHandleFunc {
	return func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		accessToken, refreshToken, err := h.auth.GetTokensFromRequest(r)
		if err != nil {
			return fmt.Errorf("cant get tokens from request in must be admin middleware func %w", err)
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
		if claims.UserID == h.cfg.AdminUserID {
		}
		return next(cart, w, r)
	}
}
