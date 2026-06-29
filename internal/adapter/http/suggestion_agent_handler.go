package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gift-app/api/internal/adapter/llmapi"
	"github.com/gift-app/api/internal/port"
)

type SuggestionAgentHandler struct {
	llmClient    *llmapi.Client
	giftRepo     port.GiftRepository
	friendRepo   port.FriendRepository
	reminderRepo port.ReminderRepository
}

type SuggestionCreateRequest struct {
	OccasionDetails string `json:"occasion_details"`
	ReminderID      string `json:"reminder_id,omitempty"`
}

type SuggestionChatRequest struct {
	GiftID  string `json:"gift_id"`
	Message string `json:"message"`
}

type SuggestionFinalizeRequest struct {
	GiftID string `json:"gift_id"`
}

func NewSuggestionAgentHandler(
	llmClient *llmapi.Client,
	giftRepo port.GiftRepository,
	friendRepo port.FriendRepository,
	reminderRepo port.ReminderRepository,
) *SuggestionAgentHandler {
	return &SuggestionAgentHandler{
		llmClient:    llmClient,
		giftRepo:     giftRepo,
		friendRepo:   friendRepo,
		reminderRepo: reminderRepo,
	}
}

// Create godoc
// @Summary     Criar sugestão inicial para friend
// @Tags        suggestion-agent
// @Accept      json
// @Produce     json
// @Param       friend_id path string true "ID do friend"
// @Param       payload body SuggestionCreateRequest true "Dados da ocasião"
// @Success     200 {object} map[string]any
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /profiles/{friend_id}/suggestions [post]
func (h *SuggestionAgentHandler) Create(w http.ResponseWriter, r *http.Request) {
	friendID := r.PathValue("friend_id")
	if friendID == "" {
		writeError(w, http.StatusBadRequest, "friend_id is required", errors.New("missing friend_id"))
		return
	}
	if !uuidPattern.MatchString(friendID) {
		writeError(w, http.StatusBadRequest, "friend_id must be a valid UUID", errors.New("invalid friend_id"))
		return
	}

	var req SuggestionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	if req.OccasionDetails == "" {
		writeError(w, http.StatusBadRequest, "occasion_details is required", errors.New("missing occasion_details"))
		return
	}
	if req.ReminderID != "" && !uuidPattern.MatchString(req.ReminderID) {
		writeError(w, http.StatusBadRequest, "reminder_id must be a valid UUID", errors.New("invalid reminder_id"))
		return
	}

	friend, err := h.friendRepo.GetByID(r.Context(), friendID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not validate friend", err)
		return
	}
	if friend == nil {
		writeError(w, http.StatusNotFound, "friend not found", errors.New("friend not found"))
		return
	}

	if req.ReminderID != "" {
		reminder, err := h.reminderRepo.GetByID(r.Context(), req.ReminderID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not validate reminder", err)
			return
		}
		if reminder == nil {
			writeError(w, http.StatusNotFound, "reminder not found", errors.New("reminder not found"))
			return
		}
		if reminder.FriendID != friendID {
			writeError(w, http.StatusConflict, "reminder does not belong to friend", errors.New("reminder-friend mismatch"))
			return
		}
	}

	response, err := h.llmClient.SuggestionCreate(r.Context(), r.Header.Get("X-Request-ID"), friendID, req)
	if err != nil {
		h.writeLLMError(w, err, "could not create initial suggestion")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// Chat godoc
// @Summary     Conversar para refinar sugestão por gift
// @Tags        suggestion-agent
// @Accept      json
// @Produce     json
// @Param       payload body SuggestionChatRequest true "Payload do chat"
// @Success     200 {object} map[string]any
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /suggestions/agent/chat [post]
func (h *SuggestionAgentHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req SuggestionChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if req.GiftID == "" || req.Message == "" {
		writeError(w, http.StatusBadRequest, "gift_id and message are required", errors.New("missing required fields"))
		return
	}

	gift, err := h.giftRepo.GetByID(r.Context(), req.GiftID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not validate gift", err)
		return
	}
	if gift == nil {
		writeError(w, http.StatusNotFound, "gift not found", errors.New("gift not found"))
		return
	}

	response, err := h.llmClient.SuggestionChat(r.Context(), r.Header.Get("X-Request-ID"), req)
	if err != nil {
		h.writeLLMError(w, err, "could not execute suggestion chat")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// Finalize godoc
// @Summary     Finalizar refinamento de sugestão por gift
// @Tags        suggestion-agent
// @Accept      json
// @Produce     json
// @Param       payload body SuggestionFinalizeRequest true "Payload de finalização"
// @Success     200 {object} map[string]any
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /suggestions/agent/finalize [post]
func (h *SuggestionAgentHandler) Finalize(w http.ResponseWriter, r *http.Request) {
	var req SuggestionFinalizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if req.GiftID == "" {
		writeError(w, http.StatusBadRequest, "gift_id is required", errors.New("missing gift_id"))
		return
	}

	gift, err := h.giftRepo.GetByID(r.Context(), req.GiftID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not validate gift", err)
		return
	}
	if gift == nil {
		writeError(w, http.StatusNotFound, "gift not found", errors.New("gift not found"))
		return
	}

	response, err := h.llmClient.SuggestionFinalize(r.Context(), r.Header.Get("X-Request-ID"), req)
	if err != nil {
		h.writeLLMError(w, err, "could not finalize suggestion session")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *SuggestionAgentHandler) writeLLMError(w http.ResponseWriter, err error, fallbackMessage string) {
	var apiErr *llmapi.APIError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode >= 400 && apiErr.StatusCode < 600 {
			writeError(w, apiErr.StatusCode, apiErr.Message, err)
			return
		}
	}
	writeError(w, http.StatusInternalServerError, fallbackMessage, err)
}
