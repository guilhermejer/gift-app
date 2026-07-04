package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gift-app/api/internal/domain"
	"github.com/gift-app/api/internal/port"
)

type ReminderHandler struct {
	repo port.ReminderRepository
}

type ReminderUpsertRequest struct {
	UserID     string                    `json:"userID" example:"a3f24e53-0d56-469d-8ea2-0dbb5f64da8a"`
	FriendID   string                    `json:"friendID" example:"9b02ce54-4f42-4a8b-a539-5b53a6e37e63"`
	Type       string                    `json:"type" example:"birthday"`
	TriggerAt  string                    `json:"triggerAt" format:"date" example:"2026-08-15"`
	Recurrence domain.ReminderRecurrence `json:"recurrence" example:"yearly"`
	Message    string                    `json:"message" example:"Comprar presente ate uma semana antes"`
}

func NewReminderHandler(repo port.ReminderRepository) *ReminderHandler {
	return &ReminderHandler{repo: repo}
}

func populateNextOccurrence(r *domain.Reminder) {
	if r == nil {
		return
	}
	if next, ok := domain.NextOccurrence(r.Recurrence, r.TriggerAt, time.Now().UTC()); ok {
		r.NextOccurrence = &next
	}
}

// Create godoc
// @Summary     Criar lembrete
// @Description Exemplo de payload: {"friendID":"9b02ce54-4f42-4a8b-a539-5b53a6e37e63","type":"birthday","triggerAt":"2026-08-15","recurrence":"yearly","message":"Comprar presente ate uma semana antes"}. Campo triggerAt no formato YYYY-MM-DD. Campo recurrence opcional (default "none"); usar "yearly"/"monthly"/"weekly"/"daily" para eventos recorrentes.
// @Tags        reminders
// @Accept      json
// @Produce     json
// @Param       userId  path string          true "ID do usuário"
// @Param       reminder body ReminderUpsertRequest true "Dados do lembrete"
// @Success     201 {object} domain.Reminder
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /users/{userId}/reminders [put]
func (h *ReminderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req ReminderUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	if req.TriggerAt == "" {
		writeError(w, http.StatusBadRequest, "trigger_at is required", errors.New("trigger_at is required"))
		return
	}

	triggerAt, err := parseDateOnly(req.TriggerAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "triggerAt must be in format YYYY-MM-DD", err)
		return
	}

	recurrence := req.Recurrence
	if recurrence == "" {
		recurrence = domain.ReminderRecurrenceNone
	}
	if !recurrence.IsValid() {
		writeError(w, http.StatusBadRequest, "recurrence must be one of: none, yearly, monthly, weekly, daily", errors.New("invalid recurrence"))
		return
	}

	reminder := domain.Reminder{
		UserID:     r.PathValue("userId"),
		FriendID:   req.FriendID,
		Type:       req.Type,
		TriggerAt:  triggerAt,
		Recurrence: recurrence,
		Message:    req.Message,
	}

	if reminder.FriendID == "" {
		writeError(w, http.StatusBadRequest, "friendID is required", errors.New("friendID is required"))
		return
	}

	created, err := h.repo.Create(r.Context(), &reminder)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create reminder", err)
		return
	}
	populateNextOccurrence(created)
	writeJSON(w, http.StatusCreated, created)
}

// Update godoc
// @Summary     Atualizar lembrete
// @Description Exemplo de payload: {"userID":"a3f24e53-0d56-469d-8ea2-0dbb5f64da8a","friendID":"9b02ce54-4f42-4a8b-a539-5b53a6e37e63","type":"anniversary","triggerAt":"2026-09-20","recurrence":"yearly","message":"Enviar flores"}. Campo triggerAt no formato YYYY-MM-DD. Atualizar recurrence recalcula as proximas ocorrencias.
// @Tags        reminders
// @Accept      json
// @Produce     json
// @Param       reminderId path string          true "ID do lembrete"
// @Param       reminder    body ReminderUpsertRequest true "Dados a atualizar"
// @Success     200 {object} domain.Reminder
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /reminders/{reminderId} [post]
func (h *ReminderHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req ReminderUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	if req.TriggerAt == "" {
		writeError(w, http.StatusBadRequest, "trigger_at is required", errors.New("trigger_at is required"))
		return
	}

	triggerAt, err := parseDateOnly(req.TriggerAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "triggerAt must be in format YYYY-MM-DD", err)
		return
	}

	recurrence := req.Recurrence
	if recurrence == "" {
		recurrence = domain.ReminderRecurrenceNone
	}
	if !recurrence.IsValid() {
		writeError(w, http.StatusBadRequest, "recurrence must be one of: none, yearly, monthly, weekly, daily", errors.New("invalid recurrence"))
		return
	}

	reminder := domain.Reminder{
		ReminderID: r.PathValue("reminderId"),
		UserID:     req.UserID,
		FriendID:   req.FriendID,
		Type:       req.Type,
		TriggerAt:  triggerAt,
		Recurrence: recurrence,
		Message:    req.Message,
	}

	updated, err := h.repo.Update(r.Context(), &reminder)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update reminder", err)
		return
	}
	if updated == nil {
		writeError(w, http.StatusNotFound, "reminder not found", errors.New("reminder not found"))
		return
	}
	populateNextOccurrence(updated)
	writeJSON(w, http.StatusOK, updated)
}

// ListByUserID godoc
// @Summary     Listar lembretes do usuário
// @Tags        reminders
// @Produce     json
// @Param       userId path string true "ID do usuário"
// @Success     200 {array}  domain.Reminder
// @Failure     500 {object} ErrorResponse
// @Router      /users/{userId}/reminders [get]
func (h *ReminderHandler) ListByUserID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	reminders, err := h.repo.ListByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list reminders", err)
		return
	}
	for _, rem := range reminders {
		populateNextOccurrence(rem)
	}
	writeJSON(w, http.StatusOK, reminders)
}
