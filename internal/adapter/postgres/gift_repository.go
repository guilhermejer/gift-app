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
	var description, priceRange, occasionDetails, reminderID *string

	giftType := gift.Type
	if !domain.IsValidGiftType(giftType) {
		giftType = domain.GiftTypeGift
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO giftowner.gifts (friend_id, title, description, price_range, tags, type, occasion_details, reminder_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING gift_id, friend_id, title, description, price_range, tags, type, occasion_details, reminder_id
	`,
		gift.FriendID,
		gift.Title,
		nullableString(gift.Description),
		nullableString(gift.PriceRange),
		gift.Tags,
		giftType,
		nullableString(gift.OccasionDetails),
		nullableString(gift.ReminderID),
	).Scan(&created.GiftID, &created.FriendID, &created.Title, &description, &priceRange, &created.Tags, &created.Type, &occasionDetails, &reminderID)
	if err != nil {
		return nil, err
	}
	if description != nil {
		created.Description = *description
	}
	if priceRange != nil {
		created.PriceRange = *priceRange
	}
	if occasionDetails != nil {
		created.OccasionDetails = *occasionDetails
	}
	if reminderID != nil {
		created.ReminderID = *reminderID
	}
	return &created, nil
}

func (r *GiftRepository) Update(ctx context.Context, gift *domain.Gift) (*domain.Gift, error) {
	var updated domain.Gift
	var description, priceRange, occasionDetails, reminderID *string

	giftType := gift.Type
	if !domain.IsValidGiftType(giftType) {
		giftType = domain.GiftTypeGift
	}

	err := r.pool.QueryRow(ctx, `
		UPDATE giftowner.gifts
		SET title = $1, description = $2, price_range = $3, tags = $4, type = $5, occasion_details = $6, reminder_id = $7, updated_at = now()
		WHERE gift_id = $8
		RETURNING gift_id, friend_id, title, description, price_range, tags, type, occasion_details, reminder_id
	`,
		gift.Title,
		nullableString(gift.Description),
		nullableString(gift.PriceRange),
		gift.Tags,
		giftType,
		nullableString(gift.OccasionDetails),
		nullableString(gift.ReminderID),
		gift.GiftID,
	).Scan(&updated.GiftID, &updated.FriendID, &updated.Title, &description, &priceRange, &updated.Tags, &updated.Type, &occasionDetails, &reminderID)
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
	if occasionDetails != nil {
		updated.OccasionDetails = *occasionDetails
	}
	if reminderID != nil {
		updated.ReminderID = *reminderID
	}
	return &updated, nil
}

func (r *GiftRepository) Delete(ctx context.Context, giftID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM giftowner.gifts WHERE gift_id = $1`, giftID)
	return err
}

func (r *GiftRepository) GetByID(ctx context.Context, giftID string) (*domain.Gift, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT gift_id, friend_id, title, description, price_range, tags, type, occasion_details, reminder_id
		FROM giftowner.gifts
		WHERE gift_id = $1
	`, giftID)

	var g domain.Gift
	var description, priceRange, occasionDetails, reminderID *string
	err := row.Scan(&g.GiftID, &g.FriendID, &g.Title, &description, &priceRange, &g.Tags, &g.Type, &occasionDetails, &reminderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if description != nil {
		g.Description = *description
	}
	if priceRange != nil {
		g.PriceRange = *priceRange
	}
	if occasionDetails != nil {
		g.OccasionDetails = *occasionDetails
	}
	if reminderID != nil {
		g.ReminderID = *reminderID
	}
	return &g, nil
}

func (r *GiftRepository) ListByFriendID(ctx context.Context, friendID string) ([]*domain.Gift, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT gift_id, friend_id, title, description, price_range, tags, type, occasion_details, reminder_id
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
		var description, priceRange, occasionDetails, reminderID *string

		if err := rows.Scan(&g.GiftID, &g.FriendID, &g.Title, &description, &priceRange, &g.Tags, &g.Type, &occasionDetails, &reminderID); err != nil {
			return nil, err
		}
		if description != nil {
			g.Description = *description
		}
		if priceRange != nil {
			g.PriceRange = *priceRange
		}
		if occasionDetails != nil {
			g.OccasionDetails = *occasionDetails
		}
		if reminderID != nil {
			g.ReminderID = *reminderID
		}
		gifts = append(gifts, &g)
	}
	return gifts, rows.Err()
}
