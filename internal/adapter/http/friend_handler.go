package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gift-app/api/internal/domain"
	"github.com/gift-app/api/internal/port"
)

type FriendHandler struct {
	repo port.FriendRepository
}

type FriendUpsertRequest struct {
	UserRelation string `json:"userRelation"`
	Name         string `json:"name"`
	Gender       string `json:"gender"`
	BirthDate    string `json:"birthDate"`
	City         string `json:"city"`
}

func NewFriendHandler(repo port.FriendRepository) *FriendHandler {
	return &FriendHandler{repo: repo}
}

// Create godoc
// @Summary     Criar amigo
// @Tags        friends
// @Accept      json
// @Produce     json
// @Param       user_id path string        true "ID do usuário"
// @Param       friend  body FriendUpsertRequest true "Dados do amigo"
// @Success     201 {object} domain.Friend
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /users/{user_id}/friends [put]
func (h *FriendHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req FriendUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	var birthDate time.Time
	if req.BirthDate != "" {
		parsedBirthDate, err := parseDateOnly(req.BirthDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "birthDate must be in format YYYY-MM-DD", err)
			return
		}
		birthDate = parsedBirthDate
	}

	friend := domain.Friend{
		UserID:       r.PathValue("user_id"),
		UserRelation: req.UserRelation,
		Name:         req.Name,
		Gender:       domain.Gender(req.Gender),
		BirthDate:    birthDate,
		City:         req.City,
	}

	if friend.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required", errors.New("name is required"))
		return
	}
	created, err := h.repo.Create(r.Context(), &friend)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create friend", err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Update godoc
// @Summary     Atualizar amigo
// @Tags        friends
// @Accept      json
// @Produce     json
// @Param       friend_id path string        true "ID do amigo"
// @Param       friend    body FriendUpsertRequest true "Dados a atualizar"
// @Success     200 {object} domain.Friend
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /friends/{friend_id} [post]
func (h *FriendHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req FriendUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	var birthDate time.Time
	if req.BirthDate != "" {
		parsedBirthDate, err := parseDateOnly(req.BirthDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "birthDate must be in format YYYY-MM-DD", err)
			return
		}
		birthDate = parsedBirthDate
	}

	friend := domain.Friend{
		FriendID:     r.PathValue("friend_id"),
		UserRelation: req.UserRelation,
		Name:         req.Name,
		Gender:       domain.Gender(req.Gender),
		BirthDate:    birthDate,
		City:         req.City,
	}

	updated, err := h.repo.Update(r.Context(), &friend)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update friend", err)
		return
	}
	if updated == nil {
		writeError(w, http.StatusNotFound, "friend not found", errors.New("friend not found"))
		return
	}
	writeJSON(w, http.StatusOK, updated)
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
		writeError(w, http.StatusInternalServerError, "could not list friends", err)
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
		writeError(w, http.StatusInternalServerError, "could not fetch friend", err)
		return
	}
	if friend == nil {
		writeError(w, http.StatusNotFound, "friend not found", errors.New("friend not found"))
		return
	}
	writeJSON(w, http.StatusOK, friend)
}
