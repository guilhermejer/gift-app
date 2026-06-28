package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/gift-app/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReminderRepository struct {
	pool *pgxpool.Pool
}

func NewReminderRepository(pool *pgxpool.Pool) *ReminderRepository {
	return &ReminderRepository{pool: pool}
}

func (r *ReminderRepository) Create(ctx context.Context, reminder *domain.Reminder) (*domain.Reminder, error) {
	var created domain.Reminder
	var msg *string
	var reminderType string

	err := r.pool.QueryRow(ctx, `
		INSERT INTO giftowner.reminders (user_id, friend_id, type, trigger_at, message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING reminder_id, user_id, friend_id, type, trigger_at, message
	`,
		reminder.UserID,
		reminder.FriendID,
		string(reminder.Type),
		reminder.TriggerAt,
		nullableString(reminder.Message),
	).Scan(&created.ReminderID, &created.UserID, &created.FriendID, &reminderType, &created.TriggerAt, &msg)
	if err != nil {
		return nil, err
	}
	created.Type = domain.ReminderType(reminderType)
	if msg != nil {
		created.Message = *msg
	}
	return &created, nil
}

func (r *ReminderRepository) Update(ctx context.Context, reminder *domain.Reminder) (*domain.Reminder, error) {
	var updated domain.Reminder
	var msg *string
	var reminderType string

	err := r.pool.QueryRow(ctx, `
		UPDATE giftowner.reminders
		SET type = $1, trigger_at = $2, message = $3, updated_at = now()
		WHERE reminder_id = $4
		RETURNING reminder_id, user_id, friend_id, type, trigger_at, message
	`,
		string(reminder.Type),
		reminder.TriggerAt,
		nullableString(reminder.Message),
		reminder.ReminderID,
	).Scan(&updated.ReminderID, &updated.UserID, &updated.FriendID, &reminderType, &updated.TriggerAt, &msg)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	updated.Type = domain.ReminderType(reminderType)
	if msg != nil {
		updated.Message = *msg
	}
	return &updated, nil
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
