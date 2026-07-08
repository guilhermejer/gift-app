package port

import (
	"context"
	"time"
)

type SuggestionJobLogRepository interface {
	Exists(ctx context.Context, reminderID string, occurrenceDate time.Time) (bool, error)
	Insert(ctx context.Context, reminderID string, occurrenceDate time.Time) error
}
