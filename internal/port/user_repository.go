package port

import (
	"context"

	"github.com/gift-app/api/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, userID string) (*domain.User, error)
}
