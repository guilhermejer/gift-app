package port

import (
	"context"
	"time"
)

type SignedURLService interface {
	UploadURL(ctx context.Context, objectName, contentType string) (string, error)
	UpdateURL(ctx context.Context, objectName, contentType string) (string, error)
	GetURL(ctx context.Context, objectName string) (string, error)
	DeleteURL(ctx context.Context, objectName string) (string, error)
	TTL() time.Duration
}
