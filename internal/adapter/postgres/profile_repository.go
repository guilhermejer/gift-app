package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

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
	var embedding any
	if len(profile.Embedding) > 0 {
		embedding = embeddingToVectorLiteral(profile.Embedding)
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO giftowner.profiles (friend_id, likes, dislikes, personality, embedding, budget)
		VALUES ($1, $2, $3, $4, $5::vector, $6)
		ON CONFLICT (friend_id) DO UPDATE
		    SET likes       = COALESCE(NULLIF(EXCLUDED.likes, '{}'), giftowner.profiles.likes),
		        dislikes    = COALESCE(NULLIF(EXCLUDED.dislikes, '{}'), giftowner.profiles.dislikes),
		        personality = COALESCE(NULLIF(EXCLUDED.personality, '{}'), giftowner.profiles.personality),
		        embedding   = COALESCE(EXCLUDED.embedding, giftowner.profiles.embedding),
		        budget      = CASE WHEN EXCLUDED.budget IS NULL THEN giftowner.profiles.budget ELSE EXCLUDED.budget END,
		        updated_at  = now()
	`, profile.FriendID, profile.Likes, profile.Dislikes, profile.Personality, embedding, profile.Budget)
	return err
}

func (r *ProfileRepository) GetByFriendID(ctx context.Context, friendID string) (*domain.Profile, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT friend_id, likes, dislikes, personality, embedding::text, budget
		FROM giftowner.profiles
		WHERE friend_id = $1
	`, friendID)

	var p domain.Profile
	var embeddingText *string
	var budget *string
	err := row.Scan(&p.FriendID, &p.Likes, &p.Dislikes, &p.Personality, &embeddingText, &budget)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if embeddingText != nil {
		embedding, parseErr := parseVectorText(*embeddingText)
		if parseErr != nil {
			return nil, parseErr
		}
		p.Embedding = embedding
	}

	p.Budget = budget

	return &p, nil
}

func embeddingToVectorLiteral(values []float32) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.FormatFloat(float64(value), 'f', -1, 32))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func parseVectorText(raw string) ([]float32, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "[")
	trimmed = strings.TrimSuffix(trimmed, "]")
	if trimmed == "" {
		return []float32{}, nil
	}

	parts := strings.Split(trimmed, ",")
	values := make([]float32, 0, len(parts))
	for _, part := range parts {
		parsed, err := strconv.ParseFloat(strings.TrimSpace(part), 32)
		if err != nil {
			return nil, fmt.Errorf("invalid embedding value: %w", err)
		}
		values = append(values, float32(parsed))
	}

	return values, nil
}
