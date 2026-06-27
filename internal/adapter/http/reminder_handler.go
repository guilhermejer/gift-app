package http

import (
	"encoding/json"
	"net/http"

	"github.com/gift-app/api/internal/domain"
	"github.com/gift-app/api/internal/port"
)

type ReminderHandler struct {
	repo port.ReminderRepository
}

func NewReminderHandler(repo port.ReminderRepository) *ReminderHandler {
	return &ReminderHandler{repo: repo}
}

// Create godoc
// @Summary     Criar lembrete
// @Tags        reminders
// @Accept      json
// @Produce     json
// @Param       user_id  path string         true "ID do usuário"
// @Param       reminder body domain.Reminder true "Dados do lembrete"
// @Success     201 {object} domain.Reminder
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /users/{user_id}/reminders [post]
func (h *ReminderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var reminder domain.Reminder
	if err := json.NewDecoder(r.Body).Decode(&reminder); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	reminder.UserID = r.PathValue("user_id")
	if reminder.ReminderID == "" {
		writeError(w, http.StatusBadRequest, "reminder_id is required")
		return
	}
	if reminder.FriendID == "" {
		writeError(w, http.StatusBadRequest, "friend_id is required")
		return
	}
	if reminder.TriggerAt.IsZero() {
		writeError(w, http.StatusBadRequest, "trigger_at is required")
		return
	}
	if err := h.repo.Create(r.Context(), &reminder); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create reminder")
		return
	}
	writeJSON(w, http.StatusCreated, reminder)
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
		writeError(w, http.StatusInternalServerError, "could not list reminders")
		return
	}
	writeJSON(w, http.StatusOK, reminders)
}
