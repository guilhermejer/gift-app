package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/gift-app/api/internal/adapter/llmapi"
	"github.com/gift-app/api/internal/port"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type ProfileAgentHandler struct {
	llmClient  *llmapi.Client
	friendRepo port.FriendRepository
}

type AgentChatRequest struct {
	FriendID string `json:"friendID" example:"9b02ce54-4f42-4a8b-a539-5b53a6e37e63"`
	Message  string `json:"message" example:"Ela prefere experiencias ou objetos?"`
}

type AgentFinalizeRequest struct {
	FriendID string `json:"friendID" example:"9b02ce54-4f42-4a8b-a539-5b53a6e37e63"`
}

func NewProfileAgentHandler(llmClient *llmapi.Client, friendRepo port.FriendRepository) *ProfileAgentHandler {
	return &ProfileAgentHandler{llmClient: llmClient, friendRepo: friendRepo}
}

// Chat godoc
// @Summary     Conversar com profile agent
// @Description Exemplo de payload: {"friendID":"9b02ce54-4f42-4a8b-a539-5b53a6e37e63","message":"Ela prefere experiencias ou objetos?"}.
// @Tags        profile-agent
// @Accept      json
// @Produce     json
// @Param       payload body AgentChatRequest true "Payload do chat"
// @Success     200 {object} map[string]any
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /profiles/agent/chat [post]
func (h *ProfileAgentHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req AgentChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if req.FriendID == "" || req.Message == "" {
		writeError(w, http.StatusBadRequest, "friendID and message are required", errors.New("missing required fields"))
		return
	}
	if !uuidPattern.MatchString(req.FriendID) {
		writeError(w, http.StatusBadRequest, "friendID must be a valid UUID", errors.New("invalid friendID"))
		return
	}

	friend, err := h.friendRepo.GetByID(r.Context(), req.FriendID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not validate friend", err)
		return
	}
	if friend == nil {
		writeError(w, http.StatusNotFound, "friend not found", errors.New("friend not found"))
		return
	}

	response, err := h.llmClient.Chat(r.Context(), r.Header.Get("X-Request-ID"), req)
	if err != nil {
		h.writeLLMError(w, err, "could not execute profile chat")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// Finalize godoc
// @Summary     Finalizar sessão de profile agent
// @Description Exemplo de payload: {"friendID":"9b02ce54-4f42-4a8b-a539-5b53a6e37e63"}.
// @Tags        profile-agent
// @Accept      json
// @Produce     json
// @Param       payload body AgentFinalizeRequest true "Payload de finalização"
// @Success     200 {object} map[string]any
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /profiles/agent/finalize [post]
func (h *ProfileAgentHandler) Finalize(w http.ResponseWriter, r *http.Request) {
	var req AgentFinalizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	if req.FriendID == "" {
		writeError(w, http.StatusBadRequest, "friendID is required", errors.New("missing friendID"))
		return
	}
	if !uuidPattern.MatchString(req.FriendID) {
		writeError(w, http.StatusBadRequest, "friendID must be a valid UUID", errors.New("invalid friendID"))
		return
	}

	friend, err := h.friendRepo.GetByID(r.Context(), req.FriendID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not validate friend", err)
		return
	}
	if friend == nil {
		writeError(w, http.StatusNotFound, "friend not found", errors.New("friend not found"))
		return
	}

	response, err := h.llmClient.Finalize(r.Context(), r.Header.Get("X-Request-ID"), req)
	if err != nil {
		h.writeLLMError(w, err, "could not finalize profile session")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// DeleteSession godoc
// @Summary     Remover sessão do profile agent
// @Tags        profile-agent
// @Produce     json
// @Param       friendId path string true "ID do friend"
// @Success     200 {object} map[string]any
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /profiles/agent/session/{friendId} [delete]
func (h *ProfileAgentHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	friendID := r.PathValue("friendId")
	if friendID == "" {
		writeError(w, http.StatusBadRequest, "friendId is required", errors.New("missing friendId"))
		return
	}
	if !uuidPattern.MatchString(friendID) {
		writeError(w, http.StatusBadRequest, "friendId must be a valid UUID", errors.New("invalid friendId"))
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

	response, err := h.llmClient.DeleteSession(r.Context(), r.Header.Get("X-Request-ID"), friendID)
	if err != nil {
		h.writeLLMError(w, err, "could not delete profile session")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *ProfileAgentHandler) writeLLMError(w http.ResponseWriter, err error, fallbackMessage string) {
	var apiErr *llmapi.APIError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode >= 400 && apiErr.StatusCode < 600 {
			writeError(w, apiErr.StatusCode, apiErr.Message, err)
			return
		}
	}
	writeError(w, http.StatusInternalServerError, fallbackMessage, err)
}
