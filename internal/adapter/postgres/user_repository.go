package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/gift-app/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO giftowner.users (user_id, active, plan_id, birth_date, city)
		VALUES ($1, $2, $3, $4, $5)
	`, user.UserID, user.Active, nullableString(user.PlanID), nullableTime(user.BirthDate), nullableString(user.City))
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT user_id, active, plan_id, birth_date, city
		FROM giftowner.users
		WHERE user_id = $1
	`, userID)

	var u domain.User
	var planID *string
	var birthDate *time.Time
	var city *string

	err := row.Scan(&u.UserID, &u.Active, &planID, &birthDate, &city)
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
