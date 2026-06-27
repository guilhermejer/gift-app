package http

import (
	"encoding/json"
	"net/http"

	"github.com/gift-app/api/internal/domain"
	"github.com/gift-app/api/internal/port"
)

type UserHandler struct {
	repo port.UserRepository
}

func NewUserHandler(repo port.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// Create godoc
// @Summary     Criar usuário
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user body domain.User true "Dados do usuário"
// @Success     201 {object} domain.User
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if user.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	if err := h.repo.Create(r.Context(), &user); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

// GetByID godoc
// @Summary     Buscar usuário
// @Tags        users
// @Produce     json
// @Param       user_id path string true "ID do usuário"
// @Success     200 {object} domain.User
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /users/{user_id} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	user, err := h.repo.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch user")
		return
	}
	if user == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, user)
}
