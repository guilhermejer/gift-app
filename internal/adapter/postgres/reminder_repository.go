package postgres

import (
	"context"
	"time"

	"github.com/gift-app/api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReminderRepository struct {
	pool *pgxpool.Pool
}

func NewReminderRepository(pool *pgxpool.Pool) *ReminderRepository {
	return &ReminderRepository{pool: pool}
}

func (r *ReminderRepository) Create(ctx context.Context, reminder *domain.Reminder) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO giftowner.reminders (reminder_id, user_id, friend_id, type, trigger_at, message)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		reminder.ReminderID,
		reminder.UserID,
		reminder.FriendID,
		string(reminder.Type),
		reminder.TriggerAt,
		nullableString(reminder.Message),
	)
	return err
}

func (r *ReminderRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.Reminder, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT reminder_id, user_id, friend_id, type, trigger_at, message
		FROM giftowner.reminders
		WHERE user_id = $1
		ORDER BY trigger_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanReminders(rows)
}

func (r *ReminderRepository) ListPending(ctx context.Context, from, to time.Time) ([]*domain.Reminder, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT reminder_id, user_id, friend_id, type, trigger_at, message
		FROM giftowner.reminders
		WHERE trigger_at >= $1 AND trigger_at <= $2
		ORDER BY trigger_at
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanReminders(rows)
}

func scanReminders(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]*domain.Reminder, error) {
	var reminders []*domain.Reminder
	for rows.Next() {
		var rem domain.Reminder
		var msg *string
		var reminderType string

		if err := rows.Scan(&rem.ReminderID, &rem.UserID, &rem.FriendID, &reminderType, &rem.TriggerAt, &msg); err != nil {
			return nil, err
		}
		rem.Type = domain.ReminderType(reminderType)
		if msg != nil {
			rem.Message = *msg
		}
		reminders = append(reminders, &rem)
	}
	return reminders, rows.Err()
}
