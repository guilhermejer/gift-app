package postgres

import (
	"context"
	"errors"

	"github.com/gift-app/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GiftRepository struct {
	pool *pgxpool.Pool
}

func NewGiftRepository(pool *pgxpool.Pool) *GiftRepository {
	return &GiftRepository{pool: pool}
}

func (r *GiftRepository) Create(ctx context.Context, gift *domain.Gift) (*domain.Gift, error) {
	var created domain.Gift
	var description, priceRange *string

	err := r.pool.QueryRow(ctx, `
		INSERT INTO giftowner.gifts (friend_id, title, description, price_range, tags)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING gift_id, friend_id, title, description, price_range, tags
	`,
		gift.FriendID,
		gift.Title,
		nullableString(gift.Description),
		nullableString(gift.PriceRange),
		gift.Tags,
	).Scan(&created.GiftID, &created.FriendID, &created.Title, &description, &priceRange, &created.Tags)
	if err != nil {
		return nil, err
	}
	if description != nil {
		created.Description = *description
	}
	if priceRange != nil {
		created.PriceRange = *priceRange
	}
	return &created, nil
}

func (r *GiftRepository) Update(ctx context.Context, gift *domain.Gift) (*domain.Gift, error) {
	var updated domain.Gift
	var description, priceRange *string

	err := r.pool.QueryRow(ctx, `
		UPDATE giftowner.gifts
		SET title = $1, description = $2, price_range = $3, tags = $4, updated_at = now()
		WHERE gift_id = $5
		RETURNING gift_id, friend_id, title, description, price_range, tags
	`,
		gift.Title,
		nullableString(gift.Description),
		nullableString(gift.PriceRange),
		gift.Tags,
		gift.GiftID,
	).Scan(&updated.GiftID, &updated.FriendID, &updated.Title, &description, &priceRange, &updated.Tags)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if description != nil {
		updated.Description = *description
	}
	if priceRange != nil {
		updated.PriceRange = *priceRange
	}
	return &updated, nil
}

func (r *GiftRepository) ListByFriendID(ctx context.Context, friendID string) ([]*domain.Gift, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT gift_id, friend_id, title, description, price_range, tags
		FROM giftowner.gifts
		WHERE friend_id = $1
		ORDER BY created_at DESC
	`, friendID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gifts []*domain.Gift
	for rows.Next() {
		var g domain.Gift
		var description, priceRange *string

		if err := rows.Scan(&g.GiftID, &g.FriendID, &g.Title, &description, &priceRange, &g.Tags); err != nil {
			return nil, err
		}
		if description != nil {
			g.Description = *description
		}
		if priceRange != nil {
			g.PriceRange = *priceRange
		}
		gifts = append(gifts, &g)
	}
	return gifts, rows.Err()
}
