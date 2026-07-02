package gcsstorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

const (
	defaultSignedURLTTL = 15 * time.Minute
)

type SignedURLService struct {
	bucketName string
	googleID   string
	privateKey []byte
	urlTTL     time.Duration
}

type serviceAccountKey struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
}

func NewSignedURLServiceFromEnv(_ context.Context) (*SignedURLService, error) {
	bucketName := strings.TrimSpace(os.Getenv("GCP_STORAGE_BUCKET_NAME"))
	if bucketName == "" {
		return nil, errors.New("GCP_STORAGE_BUCKET_NAME is not set")
	}

	credentialsPath := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if credentialsPath == "" {
		credentialsPath = strings.TrimSpace(os.Getenv("GCP_SERVICE_ACCOUNT_JSON_PATH"))
	}
	if credentialsPath == "" {
		return nil, errors.New("GOOGLE_APPLICATION_CREDENTIALS or GCP_SERVICE_ACCOUNT_JSON_PATH is not set")
	}

	keyData, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("could not read service account key file: %w", err)
	}

	var key serviceAccountKey
	if err := json.Unmarshal(keyData, &key); err != nil {
		return nil, fmt.Errorf("could not parse service account key file: %w", err)
	}
	if key.ClientEmail == "" || key.PrivateKey == "" {
		return nil, errors.New("service account key file must contain client_email and private_key")
	}

	urlTTL := defaultSignedURLTTL
	if rawTTL := strings.TrimSpace(os.Getenv("GCP_STORAGE_SIGNED_URL_TTL_SECONDS")); rawTTL != "" {
		duration, parseErr := time.ParseDuration(rawTTL + "s")
		if parseErr == nil && duration > 0 {
			urlTTL = duration
		}
	}

	return &SignedURLService{
		bucketName: bucketName,
		googleID:   key.ClientEmail,
		privateKey: []byte(key.PrivateKey),
		urlTTL:     urlTTL,
	}, nil
}

func (s *SignedURLService) UploadURL(_ context.Context, objectName, contentType string) (string, error) {
	return s.signedURL(http.MethodPut, objectName, contentType)
}

func (s *SignedURLService) UpdateURL(_ context.Context, objectName, contentType string) (string, error) {
	return s.signedURL(http.MethodPut, objectName, contentType)
}

func (s *SignedURLService) GetURL(_ context.Context, objectName string) (string, error) {
	return s.signedURL(http.MethodGet, objectName, "")
}

func (s *SignedURLService) DeleteURL(_ context.Context, objectName string) (string, error) {
	return s.signedURL(http.MethodDelete, objectName, "")
}

func (s *SignedURLService) TTL() time.Duration {
	if s == nil || s.urlTTL <= 0 {
		return defaultSignedURLTTL
	}
	return s.urlTTL
}

func (s *SignedURLService) signedURL(method, objectName, contentType string) (string, error) {
	if s == nil {
		return "", errors.New("signed url service is nil")
	}
	objectName = strings.TrimSpace(objectName)
	if objectName == "" {
		return "", errors.New("objectName is required")
	}

	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		GoogleAccessID: s.googleID,
		PrivateKey:     s.privateKey,
		Method:         method,
		Expires:        time.Now().Add(s.urlTTL),
	}
	if contentType != "" {
		opts.ContentType = contentType
	}

	url, err := storage.SignedURL(s.bucketName, objectName, opts)
	if err != nil {
		return "", err
	}
	return url, nil
}
