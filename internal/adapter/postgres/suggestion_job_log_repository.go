package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SuggestionJobLogRepository struct {
	pool *pgxpool.Pool
}

func NewSuggestionJobLogRepository(pool *pgxpool.Pool) *SuggestionJobLogRepository {
	return &SuggestionJobLogRepository{pool: pool}
}

func (r *SuggestionJobLogRepository) Exists(ctx context.Context, reminderID string, occurrenceDate time.Time) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM giftowner.suggestion_job_log
			WHERE reminder_id = $1 AND occurrence_date = $2
		)
	`, reminderID, occurrenceDate).Scan(&exists)
	return exists, err
}

func (r *SuggestionJobLogRepository) Insert(ctx context.Context, reminderID string, occurrenceDate time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO giftowner.suggestion_job_log (reminder_id, occurrence_date)
		VALUES ($1, $2)
		ON CONFLICT (reminder_id, occurrence_date) DO NOTHING
	`, reminderID, occurrenceDate)
	return err
}
