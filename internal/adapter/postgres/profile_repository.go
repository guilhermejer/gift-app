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
		INSERT INTO giftowner.profiles (friend_id, likes, dislikes, embedding)
		VALUES ($1, $2, $3, $4::vector)
		ON CONFLICT (friend_id) DO UPDATE
		    SET likes     = EXCLUDED.likes,
		        dislikes  = EXCLUDED.dislikes,
		        embedding = EXCLUDED.embedding,
		        updated_at = now()
	`, profile.FriendID, profile.Likes, profile.Dislikes, embedding)
	return err
}

func (r *ProfileRepository) GetByFriendID(ctx context.Context, friendID string) (*domain.Profile, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT friend_id, likes, dislikes, embedding::text
		FROM giftowner.profiles
		WHERE friend_id = $1
	`, friendID)

	var p domain.Profile
	var embeddingText *string
	err := row.Scan(&p.FriendID, &p.Likes, &p.Dislikes, &embeddingText)
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
