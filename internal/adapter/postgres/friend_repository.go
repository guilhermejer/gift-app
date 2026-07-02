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

func (r *FriendRepository) Create(ctx context.Context, friend *domain.Friend) (*domain.Friend, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO giftowner.friends (user_id, user_relation, name, avatar, gender, birth_date, city)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING friend_id, user_id, user_relation, name, avatar, gender, birth_date, city
	`,
		friend.UserID,
		friend.UserRelation,
		friend.Name,
		nullableString(friend.Avatar),
		string(friend.Gender),
		nullableTime(friend.BirthDate),
		nullableString(friend.City),
	)
	return scanFriend(row)
}

func (r *FriendRepository) Update(ctx context.Context, friend *domain.Friend) (*domain.Friend, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE giftowner.friends
		SET user_relation = $1, name = $2, avatar = $3, gender = $4, birth_date = $5, city = $6, updated_at = now()
		WHERE friend_id = $7
		RETURNING friend_id, user_id, user_relation, name, avatar, gender, birth_date, city
	`,
		friend.UserRelation,
		friend.Name,
		nullableString(friend.Avatar),
		string(friend.Gender),
		nullableTime(friend.BirthDate),
		nullableString(friend.City),
		friend.FriendID,
	)
	return scanFriend(row)
}

func (r *FriendRepository) GetByID(ctx context.Context, friendID string) (*domain.Friend, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT friend_id, user_id, user_relation, name, avatar, gender, birth_date, city
		FROM giftowner.friends
		WHERE friend_id = $1
	`, friendID)

	return scanFriend(row)
}

func (r *FriendRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.Friend, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT friend_id, user_id, user_relation, name, avatar, gender, birth_date, city
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
	var avatar *string
	var gender *string
	var birthDate *time.Time
	var city *string

	err := s.Scan(&f.FriendID, &f.UserID, &f.UserRelation, &f.Name, &avatar, &gender, &birthDate, &city)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if avatar != nil {
		f.Avatar = *avatar
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
