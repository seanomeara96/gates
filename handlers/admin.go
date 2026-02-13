package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repos"
	"github.com/seanomeara96/gates/views/pages"
)

func (h *Handler) AdminLogin(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("admin login: parse form: %w", err)
	}

	userID := r.Form.Get("user_id")
	accessToken, refreshToken, err := h.auth.Login(r.Context(), userID, r.Form.Get("password"))
	if err != nil {
		return fmt.Errorf("admin login: authenticate user_id=%q: %w", userID, err)
	}
	h.auth.SetTokens(w, accessToken, refreshToken)

	http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
	return nil
}

func (h *Handler) GetAdminDashboard(cart models.Cart, w http.ResponseWriter, r *http.Request) error {

	accessToken, refreshToken, err := h.auth.GetTokensFromRequest(r)
	if err != nil {
		return fmt.Errorf("admin dashboard: get tokens from request: %w", err)
	}

	_, err = h.auth.ValidateToken(accessToken)
	if err != nil {
		return fmt.Errorf("admin dashboard: validate access token: %w", err)
	}

	accessToken, refreshToken, err = h.auth.Refresh(r.Context(), refreshToken)
	if err != nil {
		return fmt.Errorf("admin dashboard: refresh tokens: %w", err)
	}
	h.auth.SetTokens(w, accessToken, refreshToken)

	orders, err := h.orderRepo.GetOrders(repos.GetOrdersParams{Limit: 25, Offset: 0})
	if err != nil {
		return fmt.Errorf("admin dashboard: fetch orders (limit=%d offset=%d): %w", 25, 0, err)
	}

	products, err := h.productRepo.GetProducts(repos.ProductFilterParams{})
	if err != nil {
		return fmt.Errorf("admin dashboard: fetch products: %w", err)
	}

	if r.URL.Query().Get("showData") == "true" {
		if err := json.NewEncoder(w).Encode(map[string]any{"Products": products, "Orders": orders}); err != nil {
			return fmt.Errorf("admin dashboard: encode response data as json: %w", err)
		}
		return nil
	}

	if h.cfg.UseTempl {
		props := pages.DashboardPageProps{
			BaseProps: pages.BaseProps{
				PageTitle: "Admin Dashboard",
				Env:       h.cfg.Mode,
				Cart:      cart,
			},
			Orders:   orders,
			Products: products,
		}
		return pages.Dashboard(props).Render(r.Context(), w)
	}

	data := map[string]any{
		"PageTitle":       "Admin Dashboard",
		"MetaDescription": "",
		"Orders":          orders,
		"Products":        products,
		"Cart":            cart,
		"Env":             h.cfg.Mode,
	}

	if err := h.rndr.Page(w, "dashboard", data); err != nil {
		return fmt.Errorf("admin dashboard: render page %q: %w", "dashboard", err)
	}
	return nil
}

func (h *Handler) GetAdminLoginPage(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	accessToken, refreshToken, _ := h.auth.GetTokensFromRequest(r)

	_, err := h.auth.ValidateToken(accessToken)
	if err == nil {
		accessToken, refreshToken, err = h.auth.Refresh(r.Context(), refreshToken)
		if err != nil {
			return fmt.Errorf("admin login page: refresh tokens for already-authenticated user: %w", err)
		}
		h.auth.SetTokens(w, accessToken, refreshToken)
		http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
		return nil
	}

	if h.cfg.UseTempl {
		props := pages.AdminLoginPageProps{
			BaseProps: pages.BaseProps{
				PageTitle: "Admin Login Page",
				Cart:      cart,
				Env:       h.cfg.Mode,
			},
		}
		return pages.AdminLogin(props).Render(r.Context(), w)
	}

	data := map[string]any{
		"PageTitle":       "Home Page",
		"MetaDescription": "Welcome to the home page",

		"Cart": cart,
		"Env":  h.cfg.Mode,
	}

	if err := h.rndr.Page(w, "admin-login", data); err != nil {
		return fmt.Errorf("admin login page: render page %q: %w", "admin-login", err)
	}
	return nil
}

func (h *Handler) Logout(cart models.Cart, w http.ResponseWriter, r *http.Request) error {

	_, refreshToken, err := h.auth.GetTokensFromRequest(r)
	if err != nil {
		log.Printf("[WARNING] Could not get tokens from request. Likely no cookie. %v", err)
	}
	if err := h.auth.Logout(r.Context(), refreshToken); err != nil {
		log.Printf("[WARNING] Could not log user out properly. %v", err)
	}

	h.auth.SetTokens(w, "", "")
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

	return nil

}

func (h *Handler) AdminLogout(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	return h.Logout(cart, w, r)
}
