package port

import (
	"context"

	"github.com/gift-app/api/internal/domain"
)

type ProfileRepository interface {
	Save(ctx context.Context, profile *domain.Profile) error
	GetByFriendID(ctx context.Context, friendID string) (*domain.Profile, error)
}
