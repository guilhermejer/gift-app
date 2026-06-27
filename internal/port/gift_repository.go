package port

import (
	"context"

	"github.com/gift-app/api/internal/domain"
)

type GiftRepository interface {
	Create(ctx context.Context, gift *domain.Gift) error
	ListByFriendID(ctx context.Context, friendID string) ([]*domain.Gift, error)
}
