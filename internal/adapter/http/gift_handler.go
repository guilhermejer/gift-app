package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gift-app/api/internal/domain"
	"github.com/gift-app/api/internal/port"
)

type GiftHandler struct {
	repo         port.GiftRepository
	friendRepo   port.FriendRepository
	reminderRepo port.ReminderRepository
}

type GiftUpsertRequest struct {
	FriendID        string   `json:"friendID"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	PriceRange      string   `json:"priceRange"`
	Tags            []string `json:"tags"`
	OccasionDetails string   `json:"occasionDetails"`
	ReminderID      string   `json:"reminderID"`
}

func NewGiftHandler(repo port.GiftRepository, friendRepo port.FriendRepository, reminderRepo port.ReminderRepository) *GiftHandler {
	return &GiftHandler{repo: repo, friendRepo: friendRepo, reminderRepo: reminderRepo}
}

// Create godoc
// @Summary     Criar sugestão de presente
// @Tags        gifts
// @Accept      json
// @Produce     json
// @Param       friend_id path string      true "ID do amigo"
// @Param       gift      body GiftUpsertRequest true "Dados do presente"
// @Success     201 {object} domain.Gift
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /friends/{friend_id}/gifts [put]
func (h *GiftHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req GiftUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	friendID := r.PathValue("friend_id")
	friend, err := h.friendRepo.GetByID(r.Context(), friendID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not validate friend", err)
		return
	}
	if friend == nil {
		writeError(w, http.StatusNotFound, "friend not found", errors.New("friend not found"))
		return
	}

	if req.FriendID != "" && req.FriendID != friendID {
		writeError(w, http.StatusConflict, "friendID in payload must match path parameter", errors.New("friendID mismatch"))
		return
	}

	gift := domain.Gift{
		FriendID:        friendID,
		Title:           req.Title,
		Description:     req.Description,
		PriceRange:      req.PriceRange,
		Tags:            req.Tags,
		OccasionDetails: req.OccasionDetails,
		ReminderID:      req.ReminderID,
	}

	if gift.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required", errors.New("title is required"))
		return
	}
	if gift.ReminderID != "" {
		reminder, err := h.reminderRepo.GetByID(r.Context(), gift.ReminderID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not validate reminder", err)
			return
		}
		if reminder == nil {
			writeError(w, http.StatusNotFound, "reminder not found", errors.New("reminder not found"))
			return
		}
		if reminder.FriendID != friendID {
			writeError(w, http.StatusConflict, "reminder does not belong to the provided friend", errors.New("reminder-friend mismatch"))
			return
		}
	}

	created, err := h.repo.Create(r.Context(), &gift)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create gift", err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Update godoc
// @Summary     Atualizar sugestão de presente
// @Tags        gifts
// @Accept      json
// @Produce     json
// @Param       gift_id path string      true "ID do presente"
// @Param       gift    body GiftUpsertRequest true "Dados a atualizar"
// @Success     200 {object} domain.Gift
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /gifts/{gift_id} [post]
func (h *GiftHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req GiftUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	giftID := r.PathValue("gift_id")
	existing, err := h.repo.GetByID(r.Context(), giftID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not validate gift", err)
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "gift not found", errors.New("gift not found"))
		return
	}

	if req.FriendID != "" && req.FriendID != existing.FriendID {
		writeError(w, http.StatusConflict, "friendID does not match the gift owner", errors.New("friendID mismatch"))
		return
	}

	gift := domain.Gift{
		GiftID:          giftID,
		FriendID:        existing.FriendID,
		Title:           req.Title,
		Description:     req.Description,
		PriceRange:      req.PriceRange,
		Tags:            req.Tags,
		OccasionDetails: existing.OccasionDetails,
		ReminderID:      existing.ReminderID,
	}
	if req.OccasionDetails != "" {
		gift.OccasionDetails = req.OccasionDetails
	}
	if req.ReminderID != "" {
		gift.ReminderID = req.ReminderID
	}

	if gift.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required", errors.New("title is required"))
		return
	}

	if gift.ReminderID != "" {
		reminder, err := h.reminderRepo.GetByID(r.Context(), gift.ReminderID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not validate reminder", err)
			return
		}
		if reminder == nil {
			writeError(w, http.StatusNotFound, "reminder not found", errors.New("reminder not found"))
			return
		}
		if reminder.FriendID != existing.FriendID {
			writeError(w, http.StatusConflict, "reminder does not belong to the gift friend", errors.New("reminder-friend mismatch"))
			return
		}
	}

	updated, err := h.repo.Update(r.Context(), &gift)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update gift", err)
		return
	}
	if updated == nil {
		writeError(w, http.StatusNotFound, "gift not found", errors.New("gift not found"))
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// ListByFriendID godoc
// @Summary     Listar sugestões de presente
// @Tags        gifts
// @Produce     json
// @Param       friend_id path string true "ID do amigo"
// @Success     200 {array}  domain.Gift
// @Failure     500 {object} map[string]string
// @Router      /friends/{friend_id}/gifts [get]
func (h *GiftHandler) ListByFriendID(w http.ResponseWriter, r *http.Request) {
	friendID := r.PathValue("friend_id")
	gifts, err := h.repo.ListByFriendID(r.Context(), friendID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list gifts", err)
		return
	}
	writeJSON(w, http.StatusOK, gifts)
}
