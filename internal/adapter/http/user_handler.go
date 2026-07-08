package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gift-app/api/internal/domain"
	"github.com/gift-app/api/internal/port"
)

type UserHandler struct {
	repo port.UserRepository
}

type UserUpsertRequest struct {
	FullName                string `json:"fullName" example:"Ana Souza"`
	Email                   string `json:"email" example:"ana.souza@email.com"`
	Active                  bool   `json:"active" example:"true"`
	PlanID                  string `json:"planId" example:"basic"`
	BirthDate               string `json:"birthDate" format:"date" example:"1992-07-21"`
	City                    string `json:"city" example:"Sao Paulo"`
	SuggestionLookaheadDays *int   `json:"suggestionLookaheadDays" example:"7"`
}

func NewUserHandler(repo port.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// Create godoc
// @Summary     Criar usuário
// @Description Exemplo de payload: {"fullName":"Ana Souza","email":"ana.souza@email.com","active":true,"planId":"basic","birthDate":"1992-07-21","city":"Sao Paulo"}. Campo birthDate no formato YYYY-MM-DD.
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user body UserUpsertRequest true "Dados do usuário"
// @Success     201 {object} domain.User
// @Failure     400 {object} ErrorResponse
// @Failure     409 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
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
	if req.SuggestionLookaheadDays != nil {
		if !domain.IsValidLookaheadDays(*req.SuggestionLookaheadDays) {
			writeError(w, http.StatusBadRequest, "suggestionLookaheadDays must be one of 7, 14, 30", errors.New("invalid lookahead"))
			return
		}
		user.SuggestionLookaheadDays = *req.SuggestionLookaheadDays
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
// @Description Exemplo de payload: {"fullName":"Ana Souza","email":"ana.novo@email.com","active":true,"planId":"pro","birthDate":"1992-07-21","city":"Campinas"}. Campo birthDate no formato YYYY-MM-DD.
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       userId path string      true "ID do usuário"
// @Param       user    body UserUpsertRequest true "Dados a atualizar"
// @Success     200 {object} domain.User
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     409 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users/{userId} [post]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req UserUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	userID := r.PathValue("userId")
	existing, err := h.repo.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch user", err)
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "user not found", errors.New("user not found"))
		return
	}

	if req.FullName != "" {
		existing.FullName = req.FullName
	}
	if req.Email != "" {
		existing.Email = req.Email
	}
	if req.BirthDate != "" {
		parsedBirthDate, err := parseDateOnly(req.BirthDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "birthDate must be in format YYYY-MM-DD", err)
			return
		}
		existing.BirthDate = parsedBirthDate
	}
	if req.PlanID != "" {
		existing.PlanID = req.PlanID
	}
	if req.City != "" {
		existing.City = req.City
	}
	if req.Active {
		existing.Active = true
	}
	if req.SuggestionLookaheadDays != nil {
		if !domain.IsValidLookaheadDays(*req.SuggestionLookaheadDays) {
			writeError(w, http.StatusBadRequest, "suggestionLookaheadDays must be one of 7, 14, 30", errors.New("invalid lookahead"))
			return
		}
		existing.SuggestionLookaheadDays = *req.SuggestionLookaheadDays
	}

	updated, err := h.repo.Update(r.Context(), existing)
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
// @Param       userId path string true "ID do usuário"
// @Success     200 {object} domain.User
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users/{userId} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
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

// GetByEmail godoc
// @Summary     Buscar usuário por email
// @Tags        users
// @Produce     json
// @Param       email query string true "Email do usuário"
// @Success     200 {object} domain.User
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users/email [get]
func (h *UserHandler) GetByEmail(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.URL.Query().Get("email"))
	if email == "" {
		writeError(w, http.StatusBadRequest, "email query parameter is required", errors.New("missing email"))
		return
	}

	user, err := h.repo.GetByEmail(r.Context(), email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch user by email", err)
		return
	}
	if user == nil {
		writeError(w, http.StatusNotFound, "user not found", errors.New("user not found"))
		return
	}

	writeJSON(w, http.StatusOK, user)
}
