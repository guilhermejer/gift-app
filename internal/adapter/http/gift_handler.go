package http

import (
	"encoding/json"
	"net/http"

	"github.com/gift-app/api/internal/domain"
	"github.com/gift-app/api/internal/port"
)

type GiftHandler struct {
	repo port.GiftRepository
}

func NewGiftHandler(repo port.GiftRepository) *GiftHandler {
	return &GiftHandler{repo: repo}
}

// Create godoc
// @Summary     Criar sugestão de presente
// @Tags        gifts
// @Accept      json
// @Produce     json
// @Param       friend_id path string      true "ID do amigo"
// @Param       gift      body domain.Gift true "Dados do presente"
// @Success     201 {object} domain.Gift
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /friends/{friend_id}/gifts [post]
func (h *GiftHandler) Create(w http.ResponseWriter, r *http.Request) {
	var gift domain.Gift
	if err := json.NewDecoder(r.Body).Decode(&gift); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	gift.FriendID = r.PathValue("friend_id")
	if gift.GiftID == "" {
		writeError(w, http.StatusBadRequest, "gift_id is required")
		return
	}
	if gift.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if err := h.repo.Create(r.Context(), &gift); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create gift")
		return
	}
	writeJSON(w, http.StatusCreated, gift)
}

// ListByFriendID godoc
// @Summary     Listar sugestões de presente
// @Tags        gifts
// @Produce     json
// @Param       friend_id path string true "ID do amigo"
// @Success     200 {array}  domain.Gift
// @Failure     500 {object} map[string]string
// @Router      /friends/{friend_id}/gifts [get]
func (h *GiftHandler) ListByFriendID(w http.ResponseWriter, r *http.Request) {
	friendID := r.PathValue("friend_id")
	gifts, err := h.repo.ListByFriendID(r.Context(), friendID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list gifts")
		return
	}
	writeJSON(w, http.StatusOK, gifts)
}
