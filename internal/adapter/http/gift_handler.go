package http

import (
	"encoding/json"
	"errors"
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
// @Router      /friends/{friend_id}/gifts [put]
func (h *GiftHandler) Create(w http.ResponseWriter, r *http.Request) {
	var gift domain.Gift
	if err := json.NewDecoder(r.Body).Decode(&gift); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	gift.FriendID = r.PathValue("friend_id")
	if gift.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required", errors.New("title is required"))
		return
	}
	created, err := h.repo.Create(r.Context(), &gift)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create gift", err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Update godoc
// @Summary     Atualizar sugestão de presente
// @Tags        gifts
// @Accept      json
// @Produce     json
// @Param       gift_id path string      true "ID do presente"
// @Param       gift    body domain.Gift true "Dados a atualizar"
// @Success     200 {object} domain.Gift
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /gifts/{gift_id} [post]
func (h *GiftHandler) Update(w http.ResponseWriter, r *http.Request) {
	var gift domain.Gift
	if err := json.NewDecoder(r.Body).Decode(&gift); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	gift.GiftID = r.PathValue("gift_id")
	updated, err := h.repo.Update(r.Context(), &gift)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update gift", err)
		return
	}
	if updated == nil {
		writeError(w, http.StatusNotFound, "gift not found", errors.New("gift not found"))
		return
	}
	writeJSON(w, http.StatusOK, updated)
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
		writeError(w, http.StatusInternalServerError, "could not list gifts", err)
		return
	}
	writeJSON(w, http.StatusOK, gifts)
}
