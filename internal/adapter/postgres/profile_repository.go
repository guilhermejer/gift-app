package postgres

import (
	"context"
	"errors"

	"github.com/gift-app/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepository struct {
	pool *pgxpool.Pool
}

func NewProfileRepository(pool *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{pool: pool}
}

func (r *ProfileRepository) Save(ctx context.Context, profile *domain.Profile) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO giftowner.profiles (friend_id, likes, dislikes)
		VALUES ($1, $2, $3)
		ON CONFLICT (friend_id) DO UPDATE
		    SET likes     = EXCLUDED.likes,
		        dislikes  = EXCLUDED.dislikes,
		        updated_at = now()
	`, profile.FriendID, profile.Likes, profile.Dislikes)
	return err
}

func (r *ProfileRepository) GetByFriendID(ctx context.Context, friendID string) (*domain.Profile, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT friend_id, likes, dislikes
		FROM giftowner.profiles
		WHERE friend_id = $1
	`, friendID)

	var p domain.Profile
	err := row.Scan(&p.FriendID, &p.Likes, &p.Dislikes)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}
