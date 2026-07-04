package port

import (
	"context"

	"github.com/gift-app/api/internal/domain"
)

type GiftRepository interface {
	Create(ctx context.Context, gift *domain.Gift) (*domain.Gift, error)
	Update(ctx context.Context, gift *domain.Gift) (*domain.Gift, error)
	Delete(ctx context.Context, giftID string) error
	GetByID(ctx context.Context, giftID string) (*domain.Gift, error)
	ListByFriendID(ctx context.Context, friendID string) ([]*domain.Gift, error)
}
