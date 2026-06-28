package port

import (
	"context"
	"time"

	"github.com/gift-app/api/internal/domain"
)

type ReminderRepository interface {
	Create(ctx context.Context, reminder *domain.Reminder) (*domain.Reminder, error)
	Update(ctx context.Context, reminder *domain.Reminder) (*domain.Reminder, error)
	ListByUserID(ctx context.Context, userID string) ([]*domain.Reminder, error)
	// ListPending retorna lembretes com trigger_at entre from e to (para envio de notificações).
	ListPending(ctx context.Context, from, to time.Time) ([]*domain.Reminder, error)
}
