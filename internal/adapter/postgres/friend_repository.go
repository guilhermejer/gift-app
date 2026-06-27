package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/gift-app/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FriendRepository struct {
	pool *pgxpool.Pool
}

func NewFriendRepository(pool *pgxpool.Pool) *FriendRepository {
	return &FriendRepository{pool: pool}
}

func (r *FriendRepository) Create(ctx context.Context, friend *domain.Friend) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO giftowner.friends (friend_id, user_id, user_relation, name, gender, birth_date, city)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		friend.FriendID,
		friend.UserID,
		friend.UserRelation,
		friend.Name,
		string(friend.Gender),
		nullableTime(friend.BirthDate),
		nullableString(friend.City),
	)
	return err
}

func (r *FriendRepository) GetByID(ctx context.Context, friendID string) (*domain.Friend, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT friend_id, user_id, user_relation, name, gender, birth_date, city
		FROM giftowner.friends
		WHERE friend_id = $1
	`, friendID)

	return scanFriend(row)
}

func (r *FriendRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.Friend, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT friend_id, user_id, user_relation, name, gender, birth_date, city
		FROM giftowner.friends
		WHERE user_id = $1
		ORDER BY name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []*domain.Friend
	for rows.Next() {
		f, err := scanFriend(rows)
		if err != nil {
			return nil, err
		}
		friends = append(friends, f)
	}
	return friends, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanFriend(s scanner) (*domain.Friend, error) {
	var f domain.Friend
	var gender *string
	var birthDate *time.Time
	var city *string

	err := s.Scan(&f.FriendID, &f.UserID, &f.UserRelation, &f.Name, &gender, &birthDate, &city)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if gender != nil {
		f.Gender = domain.Gender(*gender)
	}
	if birthDate != nil {
		f.BirthDate = *birthDate
	}
	if city != nil {
		f.City = *city
	}

	return &f, nil
}
