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
	OccasionDetails string `json:"occasionDetails" example:"Aniversario em 2026-08-15, prefere experiencias ao ar livre"`
	ReminderID      string `json:"reminderID,omitempty" example:"d8c8efdf-c52f-4d6b-8e2e-b83f78de4f77"`
}

type SuggestionChatRequest struct {
	GiftID  string `json:"giftID" example:"77a6f2b8-78f4-4cc3-bcc8-1e7f2499abf0"`
	Message string `json:"message" example:"Pode sugerir uma opcao mais economica?"`
}

type SuggestionFinalizeRequest struct {
	GiftID string `json:"giftID" example:"77a6f2b8-78f4-4cc3-bcc8-1e7f2499abf0"`
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
// @Description Exemplo de payload: {"occasionDetails":"Aniversario em 2026-08-15, prefere experiencias ao ar livre","reminderID":"d8c8efdf-c52f-4d6b-8e2e-b83f78de4f77"}.
// @Tags        suggestion-agent
// @Accept      json
// @Produce     json
// @Param       friendId path string true "ID do friend"
// @Param       payload body SuggestionCreateRequest true "Dados da ocasião"
// @Success     200 {object} map[string]any
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     409 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /profiles/{friendId}/suggestions [post]
func (h *SuggestionAgentHandler) Create(w http.ResponseWriter, r *http.Request) {
	friendID := r.PathValue("friendId")
	if friendID == "" {
		writeError(w, http.StatusBadRequest, "friendId is required", errors.New("missing friendId"))
		return
	}
	if !uuidPattern.MatchString(friendID) {
		writeError(w, http.StatusBadRequest, "friendId must be a valid UUID", errors.New("invalid friendId"))
		return
	}

	var req SuggestionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	if req.OccasionDetails == "" {
		writeError(w, http.StatusBadRequest, "occasionDetails is required", errors.New("missing occasionDetails"))
		return
	}
	if req.ReminderID != "" && !uuidPattern.MatchString(req.ReminderID) {
		writeError(w, http.StatusBadRequest, "reminderID must be a valid UUID", errors.New("invalid reminderID"))
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
// @Description Exemplo de payload: {"giftID":"77a6f2b8-78f4-4cc3-bcc8-1e7f2499abf0","message":"Pode sugerir uma opcao mais economica?"}.
// @Tags        suggestion-agent
// @Accept      json
// @Produce     json
// @Param       payload body SuggestionChatRequest true "Payload do chat"
// @Success     200 {object} map[string]any
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /suggestions/agent/chat [post]
func (h *SuggestionAgentHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req SuggestionChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if req.GiftID == "" || req.Message == "" {
		writeError(w, http.StatusBadRequest, "giftID and message are required", errors.New("missing required fields"))
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
// @Description Exemplo de payload: {"giftID":"77a6f2b8-78f4-4cc3-bcc8-1e7f2499abf0"}.
// @Tags        suggestion-agent
// @Accept      json
// @Produce     json
// @Param       payload body SuggestionFinalizeRequest true "Payload de finalização"
// @Success     200 {object} map[string]any
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /suggestions/agent/finalize [post]
func (h *SuggestionAgentHandler) Finalize(w http.ResponseWriter, r *http.Request) {
	var req SuggestionFinalizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if req.GiftID == "" {
		writeError(w, http.StatusBadRequest, "giftID is required", errors.New("missing giftID"))
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
