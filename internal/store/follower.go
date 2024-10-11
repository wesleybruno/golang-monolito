package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Follower struct {
	UserId     int64  `json:"user_id"`
	FollowerId int64  `json:"follower_Id"`
	CreatedAt  string `json:"created_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func (p FollowerStore) Follow(ctx context.Context, currentId int64, followId int64) error {
	query := `
		INSERT INTO 
			followers (user_id, follower_id ) 
		VALUES 
			($1, $2)
	`

	ctxWTimeout, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	res, err := p.db.ExecContext(ctxWTimeout, query, currentId, followId)

	if err != nil {

		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrDuplicateKey
		}

		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (p FollowerStore) Unfollow(ctx context.Context, currentId int64, followId int64) error {
	query := `
		DELETE FROM 
			followers
		WHERE 
			user_id = $1 AND follower_id = $2
		`

	ctxWTimeout, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	res, err := p.db.ExecContext(ctxWTimeout, query, currentId, followId)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
