package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gift-app/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	var created domain.User
	var planID *string
	var birthDate *time.Time
	var city *string

	lookahead := user.SuggestionLookaheadDays
	if !domain.IsValidLookaheadDays(lookahead) {
		lookahead = domain.DefaultLookaheadDays
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO giftowner.users (full_name, email, active, plan_id, birth_date, city, suggestion_lookahead_days)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING user_id, active, plan_id, birth_date, city, full_name, email, suggestion_lookahead_days
	`, user.FullName, user.Email, user.Active, nullableString(user.PlanID), nullableTime(user.BirthDate), nullableString(user.City), lookahead).
		Scan(&created.UserID, &created.Active, &planID, &birthDate, &city, &created.FullName, &created.Email, &created.SuggestionLookaheadDays)
	if err != nil {
		if isUniqueEmailViolation(err) {
			return nil, fmt.Errorf("%w: %s", domain.ErrUserEmailAlreadyExists, user.Email)
		}
		return nil, err
	}
	if planID != nil {
		created.PlanID = *planID
	}
	if birthDate != nil {
		created.BirthDate = *birthDate
	}
	if city != nil {
		created.City = *city
	}
	return &created, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	var updated domain.User
	var planID *string
	var birthDate *time.Time
	var city *string

	err := r.pool.QueryRow(ctx, `
		UPDATE giftowner.users
		SET active = $1, plan_id = $2, birth_date = $3, city = $4, full_name = $5, email = $6, suggestion_lookahead_days = $7, updated_at = now()
		WHERE user_id = $8
		RETURNING user_id, active, plan_id, birth_date, city, full_name, email, suggestion_lookahead_days
	`, user.Active, nullableString(user.PlanID), nullableTime(user.BirthDate), nullableString(user.City), user.FullName, user.Email, user.SuggestionLookaheadDays, user.UserID).
		Scan(&updated.UserID, &updated.Active, &planID, &birthDate, &city, &updated.FullName, &updated.Email, &updated.SuggestionLookaheadDays)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		if isUniqueEmailViolation(err) {
			return nil, fmt.Errorf("%w: %s", domain.ErrUserEmailAlreadyExists, user.Email)
		}
		return nil, err
	}
	if planID != nil {
		updated.PlanID = *planID
	}
	if birthDate != nil {
		updated.BirthDate = *birthDate
	}
	if city != nil {
		updated.City = *city
	}
	return &updated, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT user_id, active, plan_id, birth_date, city, full_name, email, suggestion_lookahead_days
		FROM giftowner.users
		WHERE user_id = $1
	`, userID)

	var u domain.User
	var planID *string
	var birthDate *time.Time
	var city *string

	err := row.Scan(&u.UserID, &u.Active, &planID, &birthDate, &city, &u.FullName, &u.Email, &u.SuggestionLookaheadDays)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if planID != nil {
		u.PlanID = *planID
	}
	if birthDate != nil {
		u.BirthDate = *birthDate
	}
	if city != nil {
		u.City = *city
	}

	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT user_id, active, plan_id, birth_date, city, full_name, email, suggestion_lookahead_days
		FROM giftowner.users
		WHERE lower(email) = lower($1)
	`, email)

	var u domain.User
	var planID *string
	var birthDate *time.Time
	var city *string

	err := row.Scan(&u.UserID, &u.Active, &planID, &birthDate, &city, &u.FullName, &u.Email, &u.SuggestionLookaheadDays)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if planID != nil {
		u.PlanID = *planID
	}
	if birthDate != nil {
		u.BirthDate = *birthDate
	}
	if city != nil {
		u.City = *city
	}

	return &u, nil
}

func (r *UserRepository) ListAll(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT user_id, active, plan_id, birth_date, city, full_name, email, suggestion_lookahead_days
		FROM giftowner.users
		ORDER BY full_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		var planID *string
		var birthDate *time.Time
		var city *string

		if err := rows.Scan(&u.UserID, &u.Active, &planID, &birthDate, &city, &u.FullName, &u.Email, &u.SuggestionLookaheadDays); err != nil {
			return nil, err
		}
		if planID != nil {
			u.PlanID = *planID
		}
		if birthDate != nil {
			u.BirthDate = *birthDate
		}
		if city != nil {
			u.City = *city
		}
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if users == nil {
		return []*domain.User{}, nil
	}
	return users, nil
}

func isUniqueEmailViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 23505 = unique_violation. In this repository, users create/update
		// can only realistically violate email uniqueness.
		if pgErr.Code == "23505" {
			return true
		}
	}

	// Fallback for wrapped/localized errors where pgconn.PgError may not be exposed.
	errMsg := err.Error()
	return strings.Contains(errMsg, "SQLSTATE 23505") || strings.Contains(errMsg, "duplicate key value violates unique constraint")
}
