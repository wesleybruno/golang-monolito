package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Password  password `json:"-"`
	Email     string   `json:"email"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

type UserStore struct {
	db *sql.DB
}

func (p *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `INSERT INTO users (username, password, email ) VALUES ($1, $2, $3) RETURNING id, created_at`

	ctxWTimeout, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	err := tx.QueryRowContext(ctxWTimeout, query, user.Username, user.Password.hash, user.Email).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}

	}

	return nil
}

func (p *UserStore) GetById(ctx context.Context, id int64) (*User, error) {
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

func (p *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, invitationExp time.Duration, userID int64) error {
	query := `INSERT INTO user_invitation (token, user_id, expiry ) VALUES ($1, $2, $3)`

	ctxWTimeout, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	_, err := tx.ExecContext(ctxWTimeout, query, token, userID, time.Now().Add(invitationExp))

	if err != nil {
		return err
	}

	return nil
}

func (p *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	return withTx(p.db, ctx, func(tx *sql.Tx) error {

		if err := p.Create(ctx, tx, user); err != nil {
			return err
		}

		if err := p.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err
		}

		return nil

	})

}

func (p *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(p.db, ctx, func(tx *sql.Tx) error {
		// 1. find the user that this token belongs to
		user, err := p.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		// 2. update the user
		user.IsActive = true
		if err := p.update(ctx, tx, user); err != nil {
			return err
		}

		// 3. clean the invitations
		if err := p.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.is_active
		FROM users u
		JOIN user_invitation ui ON u.id = ui.user_id
		WHERE ui.token = $1 AND ui.expiry > $2
	`

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	user := &User{}
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `UPDATE users SET username = $1, email = $2, is_active = $3 WHERE id = $4`

	ctx, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitation WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}
