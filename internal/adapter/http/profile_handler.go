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
	FriendID    string    `json:"friendID" example:"9b02ce54-4f42-4a8b-a539-5b53a6e37e63"`
	Likes       []string  `json:"likes,omitempty" example:"fotografia,musica,viagem"`
	Dislikes    []string  `json:"dislikes,omitempty" example:"multidoes,atraso"`
	Personality []string  `json:"personality,omitempty" example:"introvertida,criativa,detalhista"`
	Embedding   []float32 `json:"embedding,omitempty" example:"0.12,-0.44,0.91"`
	Budget      *string   `json:"budget,omitempty" example:"Até R$100"`
}

func NewProfileHandler(repo port.ProfileRepository) *ProfileHandler {
	return &ProfileHandler{repo: repo}
}

// Save godoc
// @Summary     Criar ou atualizar perfil do amigo
// @Description Exemplo de payload: {"friendID":"9b02ce54-4f42-4a8b-a539-5b53a6e37e63","likes":["fotografia","musica","viagem"],"dislikes":["multidoes","atraso"],"personality":["introvertida","criativa"],"embedding":[0.12,-0.44,0.91]}.
// @Tags        profiles
// @Accept      json
// @Produce     json
// @Param       friendId path  string         true "ID do amigo"
// @Param       profile   body  ProfileUpsertRequest true "Dados do perfil"
// @Success     200 {object} domain.Profile
// @Failure     400 {object} ErrorResponse
// @Failure     422 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /friends/{friendId}/profile [put]
func (h *ProfileHandler) Save(w http.ResponseWriter, r *http.Request) {
	var req ProfileUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	pathFriendID := r.PathValue("friendId")
	if req.FriendID == "" {
		req.FriendID = pathFriendID
	}
	if req.FriendID != pathFriendID {
		writeError(w, http.StatusUnprocessableEntity, "friendID in body must match path parameter", errors.New("friendID mismatch"))
		return
	}

	profile := domain.Profile{
		FriendID:    req.FriendID,
		Likes:       req.Likes,
		Dislikes:    req.Dislikes,
		Personality: req.Personality,
		Embedding:   req.Embedding,
		Budget:      req.Budget,
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
// @Param       friendId path string true "ID do amigo"
// @Success     200 {object} domain.Profile
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /friends/{friendId}/profile [get]
func (h *ProfileHandler) GetByFriendID(w http.ResponseWriter, r *http.Request) {
	friendID := r.PathValue("friendId")
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
