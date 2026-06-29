package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gift-app/api/internal/domain"
	"github.com/gift-app/api/internal/port"
)

type ProfileHandler struct {
	repo port.ProfileRepository
}

type ProfileUpsertRequest struct {
	FriendID  string    `json:"friend_id"`
	Likes     []string  `json:"likes"`
	Dislikes  []string  `json:"dislikes"`
	Embedding []float32 `json:"embedding"`
}

func NewProfileHandler(repo port.ProfileRepository) *ProfileHandler {
	return &ProfileHandler{repo: repo}
}

// Save godoc
// @Summary     Criar ou atualizar perfil do amigo
// @Tags        profiles
// @Accept      json
// @Produce     json
// @Param       friend_id path  string         true "ID do amigo"
// @Param       profile   body  ProfileUpsertRequest true "Dados do perfil"
// @Success     200 {object} domain.Profile
// @Failure     400 {object} map[string]string
// @Failure     422 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /friends/{friend_id}/profile [put]
func (h *ProfileHandler) Save(w http.ResponseWriter, r *http.Request) {
	var req ProfileUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	pathFriendID := r.PathValue("friend_id")
	if req.FriendID == "" {
		req.FriendID = pathFriendID
	}
	if req.FriendID != pathFriendID {
		writeError(w, http.StatusUnprocessableEntity, "friend_id in body must match path parameter", errors.New("friend_id mismatch"))
		return
	}

	profile := domain.Profile{
		FriendID:  req.FriendID,
		Likes:     req.Likes,
		Dislikes:  req.Dislikes,
		Embedding: req.Embedding,
	}

	if err := h.repo.Save(r.Context(), &profile); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save profile", err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

// GetByFriendID godoc
// @Summary     Buscar perfil do amigo
// @Tags        profiles
// @Produce     json
// @Param       friend_id path string true "ID do amigo"
// @Success     200 {object} domain.Profile
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /friends/{friend_id}/profile [get]
func (h *ProfileHandler) GetByFriendID(w http.ResponseWriter, r *http.Request) {
	friendID := r.PathValue("friend_id")
	profile, err := h.repo.GetByFriendID(r.Context(), friendID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch profile", err)
		return
	}
	if profile == nil {
		writeError(w, http.StatusNotFound, "profile not found", errors.New("profile not found"))
		return
	}
	writeJSON(w, http.StatusOK, profile)
}
