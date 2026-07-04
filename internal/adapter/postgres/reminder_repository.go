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

	err := r.pool.QueryRow(ctx, `
		INSERT INTO giftowner.reminders (user_id, friend_id, type, trigger_at, recurrence, message)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING reminder_id, user_id, friend_id, type, trigger_at, recurrence, message
	`,
		reminder.UserID,
		reminder.FriendID,
		reminder.Type,
		reminder.TriggerAt,
		string(reminder.Recurrence),
		nullableString(reminder.Message),
	).Scan(&created.ReminderID, &created.UserID, &created.FriendID, &created.Type, &created.TriggerAt, &created.Recurrence, &msg)
	if err != nil {
		return nil, err
	}

	if msg != nil {
		created.Message = *msg
	}
	return &created, nil
}

func (r *ReminderRepository) Update(ctx context.Context, reminder *domain.Reminder) (*domain.Reminder, error) {
	var updated domain.Reminder
	var msg *string

	err := r.pool.QueryRow(ctx, `
		UPDATE giftowner.reminders
		SET type = $1, trigger_at = $2, recurrence = $3, message = $4, updated_at = now()
		WHERE reminder_id = $5
		RETURNING reminder_id, user_id, friend_id, type, trigger_at, recurrence, message
	`,
		reminder.Type,
		reminder.TriggerAt,
		string(reminder.Recurrence),
		nullableString(reminder.Message),
		reminder.ReminderID,
	).Scan(&updated.ReminderID, &updated.UserID, &updated.FriendID, &updated.Type, &updated.TriggerAt, &updated.Recurrence, &msg)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if msg != nil {
		updated.Message = *msg
	}
	return &updated, nil
}

func (r *ReminderRepository) GetByID(ctx context.Context, reminderID string) (*domain.Reminder, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT reminder_id, user_id, friend_id, type, trigger_at, recurrence, message
		FROM giftowner.reminders
		WHERE reminder_id = $1
	`, reminderID)

	var reminder domain.Reminder
	var message *string

	err := row.Scan(&reminder.ReminderID, &reminder.UserID, &reminder.FriendID, &reminder.Type, &reminder.TriggerAt, &reminder.Recurrence, &message)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if message != nil {
		reminder.Message = *message
	}

	return &reminder, nil
}

func (r *ReminderRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.Reminder, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT reminder_id, user_id, friend_id, type, trigger_at, recurrence, message
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
		SELECT reminder_id, user_id, friend_id, type, trigger_at, recurrence, message
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

		if err := rows.Scan(&rem.ReminderID, &rem.UserID, &rem.FriendID, &rem.Type, &rem.TriggerAt, &rem.Recurrence, &msg); err != nil {
			return nil, err
		}
		if msg != nil {
			rem.Message = *msg
		}
		reminders = append(reminders, &rem)
	}
	return reminders, rows.Err()
}
