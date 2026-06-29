package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gift-app/api/internal/domain"
	"github.com/gift-app/api/internal/port"
)

type ReminderHandler struct {
	repo port.ReminderRepository
}

type ReminderUpsertRequest struct {
	UserID    string `json:"userID"`
	FriendID  string `json:"friendID"`
	Type      string `json:"type"`
	TriggerAt string `json:"triggerAt"`
	Message   string `json:"message"`
}

func NewReminderHandler(repo port.ReminderRepository) *ReminderHandler {
	return &ReminderHandler{repo: repo}
}

// Create godoc
// @Summary     Criar lembrete
// @Tags        reminders
// @Accept      json
// @Produce     json
// @Param       user_id  path string          true "ID do usuário"
// @Param       reminder body ReminderUpsertRequest true "Dados do lembrete"
// @Success     201 {object} domain.Reminder
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /users/{user_id}/reminders [put]
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

	reminder := domain.Reminder{
		UserID:    r.PathValue("user_id"),
		FriendID:  req.FriendID,
		Type:      req.Type,
		TriggerAt: triggerAt,
		Message:   req.Message,
	}

	if reminder.FriendID == "" {
		writeError(w, http.StatusBadRequest, "friend_id is required", errors.New("friend_id is required"))
		return
	}

	created, err := h.repo.Create(r.Context(), &reminder)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create reminder", err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Update godoc
// @Summary     Atualizar lembrete
// @Tags        reminders
// @Accept      json
// @Produce     json
// @Param       reminder_id path string          true "ID do lembrete"
// @Param       reminder    body ReminderUpsertRequest true "Dados a atualizar"
// @Success     200 {object} domain.Reminder
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /reminders/{reminder_id} [post]
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

	reminder := domain.Reminder{
		ReminderID: r.PathValue("reminder_id"),
		UserID:     req.UserID,
		FriendID:   req.FriendID,
		Type:       req.Type,
		TriggerAt:  triggerAt,
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
	writeJSON(w, http.StatusOK, updated)
}

// ListByUserID godoc
// @Summary     Listar lembretes do usuário
// @Tags        reminders
// @Produce     json
// @Param       user_id path string true "ID do usuário"
// @Success     200 {array}  domain.Reminder
// @Failure     500 {object} map[string]string
// @Router      /users/{user_id}/reminders [get]
func (h *ReminderHandler) ListByUserID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	reminders, err := h.repo.ListByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list reminders", err)
		return
	}
	writeJSON(w, http.StatusOK, reminders)
}
