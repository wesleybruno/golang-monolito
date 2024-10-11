package store

import (
	"context"
	"database/sql"
)

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"-"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type UserStore struct {
	db *sql.DB
}

func (p UserStore) Create(ctx context.Context, user *User) error {
	query := `INSERT INTO users (username, password, email ) VALUES ($1, $2, $3) RETURNING id, created_at`

	ctxWTimeout, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	err := p.db.QueryRowContext(ctxWTimeout, query, user.Username, user.Password, user.Email).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (p UserStore) GetById(ctx context.Context, id int64) (*User, error) {
	query := `SELECT id, username, email, created_at
		FROM users
		WHERE id = $1`

	ctxWTimeout, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	var user User

	err := p.db.QueryRowContext(ctxWTimeout, query, id).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
