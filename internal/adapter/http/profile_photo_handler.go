package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gift-app/api/internal/port"
)

type ProfilePhotoHandler struct {
	friendRepo port.FriendRepository
	storage    port.SignedURLService
}

type ProfilePhotoSignedURLRequest struct {
	ObjectName  string `json:"objectName,omitempty" example:"friends/9b02ce54-4f42-4a8b-a539-5b53a6e37e63/profile-photo"`
	ContentType string `json:"contentType,omitempty" example:"image/jpeg"`
}

type ProfilePhotoSignedURLResponse struct {
	FriendID   string    `json:"friendId" example:"9b02ce54-4f42-4a8b-a539-5b53a6e37e63"`
	ObjectName string    `json:"objectName" example:"friends/9b02ce54-4f42-4a8b-a539-5b53a6e37e63/profile-photo"`
	Method     string    `json:"method" example:"PUT"`
	URL        string    `json:"url" example:"https://storage.googleapis.com/..."`
	ExpiresAt  time.Time `json:"expiresAt" format:"date-time" example:"2026-07-01T12:30:00Z"`
}

func NewProfilePhotoHandler(friendRepo port.FriendRepository, storage port.SignedURLService) *ProfilePhotoHandler {
	return &ProfilePhotoHandler{friendRepo: friendRepo, storage: storage}
}

// CreateUploadURL godoc
// @Summary     Gerar URL assinada para upload da foto do amigo
// @Description O front deve usar a URL retornada com o metodo PUT. Se objectName nao for enviado, usa friends/{friendId}/profile-photo.
// @Tags        profile-photos
// @Accept      json
// @Produce     json
// @Param       friendId path string true "ID do amigo"
// @Param       payload body ProfilePhotoSignedURLRequest true "Dados do arquivo"
// @Success     200 {object} ProfilePhotoSignedURLResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /friends/{friendId}/profile-photo/upload-url [post]
func (h *ProfilePhotoHandler) CreateUploadURL(w http.ResponseWriter, r *http.Request) {
	h.generateURL(w, r, http.MethodPut, false)
}

// CreateUpdateURL godoc
// @Summary     Gerar URL assinada para atualizar a foto do amigo
// @Description O front deve usar a URL retornada com o metodo PUT. Se objectName nao for enviado, usa friends/{friendId}/profile-photo.
// @Tags        profile-photos
// @Accept      json
// @Produce     json
// @Param       friendId path string true "ID do amigo"
// @Param       payload body ProfilePhotoSignedURLRequest true "Dados do arquivo"
// @Success     200 {object} ProfilePhotoSignedURLResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /friends/{friendId}/profile-photo/update-url [post]
func (h *ProfilePhotoHandler) CreateUpdateURL(w http.ResponseWriter, r *http.Request) {
	h.generateURL(w, r, http.MethodPut, true)
}

// CreateDeleteURL godoc
// @Summary     Gerar URL assinada para remover a foto do amigo
// @Description O front deve usar a URL retornada com o metodo DELETE. Se objectName nao for enviado, usa friends/{friendId}/profile-photo.
// @Tags        profile-photos
// @Accept      json
// @Produce     json
// @Param       friendId path string true "ID do amigo"
// @Success     200 {object} ProfilePhotoSignedURLResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /friends/{friendId}/profile-photo/delete-url [post]
func (h *ProfilePhotoHandler) CreateDeleteURL(w http.ResponseWriter, r *http.Request) {
	h.generateURL(w, r, http.MethodDelete, false)
}

// CreateGetURL godoc
// @Summary     Gerar URL assinada para leitura da foto do amigo
// @Description O front deve usar a URL retornada com o metodo GET para obter a foto. Se objectName nao for enviado, usa friends/{friendId}/profile-photo.jpg.
// @Tags        profile-photos
// @Accept      json
// @Produce     json
// @Param       friendId path string true "ID do amigo"
// @Param       payload body ProfilePhotoSignedURLRequest false "Opcional: objectName especifico"
// @Success     200 {object} ProfilePhotoSignedURLResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /friends/{friendId}/profile-photo/get-url [post]
func (h *ProfilePhotoHandler) CreateGetURL(w http.ResponseWriter, r *http.Request) {
	h.generateURL(w, r, http.MethodGet, false)
}

func (h *ProfilePhotoHandler) generateURL(w http.ResponseWriter, r *http.Request, method string, isUpdate bool) {
	friendID := r.PathValue("friendId")
	if friendID == "" {
		writeError(w, http.StatusBadRequest, "friendId is required", errors.New("missing friendId"))
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

	var req ProfilePhotoSignedURLRequest
	if method != http.MethodDelete {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if errors.Is(err, io.EOF) {
				req = ProfilePhotoSignedURLRequest{}
			} else {
				writeError(w, http.StatusBadRequest, "invalid request body", err)
				return
			}
		}
	}

	objectName := strings.TrimSpace(req.ObjectName)
	if objectName == "" {
		objectName = fmt.Sprintf("friends/%s/profile-photo.jpg", friendID)
	}

	if method != http.MethodDelete && strings.TrimSpace(req.ContentType) == "" {
		writeError(w, http.StatusBadRequest, "contentType is required", errors.New("missing contentType"))
		return
	}

	var signedURL string
	switch method {
	case http.MethodPut:
		if isUpdate {
			signedURL, err = h.storage.UpdateURL(r.Context(), objectName, req.ContentType)
		} else {
			signedURL, err = h.storage.UploadURL(r.Context(), objectName, req.ContentType)
		}
	case http.MethodGet:
		signedURL, err = h.storage.GetURL(r.Context(), objectName)
	case http.MethodDelete:
		signedURL, err = h.storage.DeleteURL(r.Context(), objectName)
	default:
		writeError(w, http.StatusInternalServerError, "unsupported method", errors.New("unsupported method"))
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate signed url", err)
		return
	}

	writeJSON(w, http.StatusOK, ProfilePhotoSignedURLResponse{
		FriendID:   friendID,
		ObjectName: objectName,
		Method:     method,
		URL:        signedURL,
		ExpiresAt:  time.Now().Add(h.storage.TTL()),
	})
}
