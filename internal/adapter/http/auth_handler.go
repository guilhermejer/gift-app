package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/gift-app/api/internal/domain"
	"github.com/gift-app/api/internal/port"
	"google.golang.org/api/idtoken"
)

type AuthHandler struct {
	userRepo       port.UserRepository
	googleClientID string
}

type AuthGoogleRequest struct {
	IDToken string `json:"idToken"`
}

func NewAuthHandler(userRepo port.UserRepository) *AuthHandler {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	return &AuthHandler{userRepo: userRepo, googleClientID: clientID}
}

// Google godoc
// @Summary     Autenticar com Google
// @Description Recebe ID token do Google, valida, faz upsert do usuário por email e retorna o user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body AuthGoogleRequest true "ID token do Google"
// @Success     200 {object} domain.User
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /auth/google [post]
func (h *AuthHandler) Google(w http.ResponseWriter, r *http.Request) {
	if h.googleClientID == "" {
		writeError(w, http.StatusInternalServerError, "GOOGLE_CLIENT_ID not configured", errors.New("missing env"))
		return
	}

	var req AuthGoogleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if req.IDToken == "" {
		writeError(w, http.StatusBadRequest, "idToken is required", errors.New("idToken is required"))
		return
	}

	payload, err := idtoken.Validate(r.Context(), req.IDToken, h.googleClientID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid Google ID token", err)
		return
	}

	email, _ := payload.Claims["email"].(string)
	if email == "" {
		writeError(w, http.StatusBadRequest, "email not found in token", errors.New("email not found"))
		return
	}

	name, _ := payload.Claims["name"].(string)

	existing, err := h.userRepo.GetByEmail(r.Context(), email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not look up user", err)
		return
	}

	if existing != nil {
		writeJSON(w, http.StatusOK, existing)
		return
	}

	user, err := h.userRepo.Create(r.Context(), &domain.User{
		FullName:                 name,
		Email:                    email,
		SuggestionLookaheadDays: domain.DefaultLookaheadDays,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create user", err)
		return
	}
	if user == nil {
		writeError(w, http.StatusInternalServerError, "user creation returned nil", errors.New("nil user"))
		return
	}

	writeJSON(w, http.StatusOK, user)
}
