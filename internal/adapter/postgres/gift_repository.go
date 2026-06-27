package postgres

import (
	"context"

	"github.com/gift-app/api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GiftRepository struct {
	pool *pgxpool.Pool
}

func NewGiftRepository(pool *pgxpool.Pool) *GiftRepository {
	return &GiftRepository{pool: pool}
}

func (r *GiftRepository) Create(ctx context.Context, gift *domain.Gift) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO giftowner.gifts (gift_id, friend_id, title, description, price_range, tags)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		gift.GiftID,
		gift.FriendID,
		gift.Title,
		nullableString(gift.Description),
		nullableString(gift.PriceRange),
		gift.Tags,
	)
	return err
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
