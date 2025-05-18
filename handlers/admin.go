package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repos"
)

func (h *Handler) AdminLogin(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form during admin login %w", err)
	}

	accessToken, refreshToken, err := h.auth.Login(r.Context(), r.Form.Get("user_id"), r.Form.Get("password"))
	if err != nil {
		return fmt.Errorf("error during admin login %w", err)
	}
	h.auth.SetTokens(w, accessToken, refreshToken)

	http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
	return nil
}

func (h *Handler) GetAdminDashboard(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {

	accessToken, refreshToken, err := h.auth.GetTokensFromRequest(r)
	if err != nil {
		return fmt.Errorf("failed to get tokens from request while getting admin dashboard %w", err)
	}

	_, err = h.auth.ValidateToken(accessToken)
	if err != nil {
		return fmt.Errorf("failed to validate access token while getting admin dashboard %w", err)
	}

	accessToken, refreshToken, err = h.auth.Refresh(r.Context(), refreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh tokens while getting admin dashboard %w", err)
	}
	h.auth.SetTokens(w, accessToken, refreshToken)

	orders, err := h.orderRepo.GetOrders(repos.GetOrdersParams{Limit: 25, Offset: 0})
	if err != nil {
		return fmt.Errorf("failed to fetch orders for the admin dashbaord %w", err)
	}

	products, err := h.productRepo.GetProducts(repos.ProductFilterParams{})
	if err != nil {
		return fmt.Errorf("failed to fetch products for admin dasboard %w", err)
	}

	data := map[string]any{
		"PageTitle":       "Admin Dashboard",
		"MetaDescription": "",
		"Orders":          orders,
		"Products":        products,
		"Cart":            cart,
		"Env":             h.cfg.Mode,
	}

	if r.URL.Query().Get("showData") == "true" {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			return err
		}
		return nil
	}
	return h.rndr.Page(w, "dashboard", data)
}

func (h *Handler) GetAdminLoginPage(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	accessToken, refreshToken, _ := h.auth.GetTokensFromRequest(r)

	_, err := h.auth.ValidateToken(accessToken)
	if err == nil {
		accessToken, refreshToken, err = h.auth.Refresh(r.Context(), refreshToken)
		if err != nil {
			return err
		}
		h.auth.SetTokens(w, accessToken, refreshToken)
		http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
		return nil
	}

	data := map[string]any{
		"PageTitle":       "Home Page",
		"MetaDescription": "Welcome to the home page",

		"Cart": cart,
		"Env":  h.cfg.Mode,
	}

	return h.rndr.Page(w, "admin-login", data)
}

func (h *Handler) AdminLogout(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {

	_, refreshToken, err := h.auth.GetTokensFromRequest(r)
	if err != nil {
		return err
	}

	if err := h.auth.Logout(r.Context(), refreshToken); err != nil {
		return err
	}
	h.auth.SetTokens(w, "", "")

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

	return nil

}
