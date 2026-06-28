package port

import (
	"context"

	"github.com/gift-app/api/internal/domain"
)

type FriendRepository interface {
	Create(ctx context.Context, friend *domain.Friend) (*domain.Friend, error)
	Update(ctx context.Context, friend *domain.Friend) (*domain.Friend, error)
	GetByID(ctx context.Context, friendID string) (*domain.Friend, error)
	ListByUserID(ctx context.Context, userID string) ([]*domain.Friend, error)
}
