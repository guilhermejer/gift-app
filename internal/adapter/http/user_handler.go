package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gift-app/api/internal/domain"
	"github.com/gift-app/api/internal/port"
)

type UserHandler struct {
	repo port.UserRepository
}

type UserUpsertRequest struct {
	FullName  string `json:"fullName"`
	Email     string `json:"email"`
	Active    bool   `json:"active"`
	PlanID    string `json:"planId"`
	BirthDate string `json:"birthDate"`
	City      string `json:"city"`
}

func NewUserHandler(repo port.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// Create godoc
// @Summary     Criar usuário
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user body UserUpsertRequest true "Dados do usuário"
// @Success     201 {object} domain.User
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /users [put]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req UserUpsertRequest
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

	user := domain.User{
		FullName:  req.FullName,
		Email:     req.Email,
		Active:    req.Active,
		PlanID:    req.PlanID,
		BirthDate: birthDate,
		City:      req.City,
	}

	created, err := h.repo.Create(r.Context(), &user)
	if err != nil {
		if errors.Is(err, domain.ErrUserEmailAlreadyExists) {
			writeError(w, http.StatusConflict, "email already exists", err)
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create user", err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Update godoc
// @Summary     Atualizar usuário
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user_id path string      true "ID do usuário"
// @Param       user    body UserUpsertRequest true "Dados a atualizar"
// @Success     200 {object} domain.User
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /users/{user_id} [post]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req UserUpsertRequest
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

	user := domain.User{
		UserID:    r.PathValue("user_id"),
		FullName:  req.FullName,
		Email:     req.Email,
		Active:    req.Active,
		PlanID:    req.PlanID,
		BirthDate: birthDate,
		City:      req.City,
	}

	updated, err := h.repo.Update(r.Context(), &user)
	if err != nil {
		if errors.Is(err, domain.ErrUserEmailAlreadyExists) {
			writeError(w, http.StatusConflict, "email already exists", err)
			return
		}
		writeError(w, http.StatusInternalServerError, "could not update user", err)
		return
	}
	if updated == nil {
		writeError(w, http.StatusNotFound, "user not found", errors.New("user not found"))
		return
	}
	writeJSON(w, http.StatusOK, updated)
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
		writeError(w, http.StatusInternalServerError, "could not fetch user", err)
		return
	}
	if user == nil {
		writeError(w, http.StatusNotFound, "user not found", errors.New("user not found"))
		return
	}
	writeJSON(w, http.StatusOK, user)
}
