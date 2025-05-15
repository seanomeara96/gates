package handlers

import (
	"net/http"

	"github.com/seanomeara96/gates/models"
)

func (h *Handler) AdminLogin(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	accessToken, refreshToken, err := h.auth.Login(r.Context(), r.Form.Get("user_id"), r.Form.Get("password"))
	if err != nil {
		return err
	}
	h.auth.SetTokens(w, accessToken, refreshToken)

	http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
	return nil
}

func (h *Handler) GetAdminDashboard(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {

	accessToken, refreshToken, err := h.auth.GetTokensFromRequest(r)
	if err != nil {
		return err
	}

	_, err = h.auth.ValidateToken(accessToken)
	if err != nil {
		return err
	}

	accessToken, refreshToken, err = h.auth.Refresh(r.Context(), refreshToken)
	if err != nil {
		return err
	}
	h.auth.SetTokens(w, accessToken, refreshToken)

	data := map[string]any{
		"PageTitle":       "Home Page",
		"MetaDescription": "Welcome to the home page",

		"Cart": cart,
		"Env":  h.cfg.Mode,
	}

	return h.renderPage(w, "home", data)
}
