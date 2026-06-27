package http

import (
	"encoding/json"
	"net/http"

	"github.com/gift-app/api/internal/domain"
	"github.com/gift-app/api/internal/port"
)

type FriendHandler struct {
	repo port.FriendRepository
}

func NewFriendHandler(repo port.FriendRepository) *FriendHandler {
	return &FriendHandler{repo: repo}
}

// Create godoc
// @Summary     Criar amigo
// @Tags        friends
// @Accept      json
// @Produce     json
// @Param       user_id path string true "ID do usuário"
// @Param       friend body domain.Friend true "Dados do amigo"
// @Success     201 {object} domain.Friend
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /users/{user_id}/friends [post]
func (h *FriendHandler) Create(w http.ResponseWriter, r *http.Request) {
	var friend domain.Friend
	if err := json.NewDecoder(r.Body).Decode(&friend); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	friend.UserID = r.PathValue("user_id")
	if friend.FriendID == "" {
		writeError(w, http.StatusBadRequest, "friend_id is required")
		return
	}
	if friend.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if err := h.repo.Create(r.Context(), &friend); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create friend")
		return
	}
	writeJSON(w, http.StatusCreated, friend)
}

// ListByUserID godoc
// @Summary     Listar amigos do usuário
// @Tags        friends
// @Produce     json
// @Param       user_id path string true "ID do usuário"
// @Success     200 {array}  domain.Friend
// @Failure     500 {object} map[string]string
// @Router      /users/{user_id}/friends [get]
func (h *FriendHandler) ListByUserID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	friends, err := h.repo.ListByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list friends")
		return
	}
	writeJSON(w, http.StatusOK, friends)
}

// GetByID godoc
// @Summary     Buscar amigo
// @Tags        friends
// @Produce     json
// @Param       friend_id path string true "ID do amigo"
// @Success     200 {object} domain.Friend
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /friends/{friend_id} [get]
func (h *FriendHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	friendID := r.PathValue("friend_id")
	friend, err := h.repo.GetByID(r.Context(), friendID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch friend")
		return
	}
	if friend == nil {
		writeError(w, http.StatusNotFound, "friend not found")
		return
	}
	writeJSON(w, http.StatusOK, friend)
}
