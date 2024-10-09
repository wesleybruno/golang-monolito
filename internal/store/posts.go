package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserId    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Comment   []Comment `json:"comments"`
	Version   int64     `json:"version"`
}

type PostStore struct {
	db *sql.DB
}

func (p PostStore) Create(ctx context.Context, post *Post) error {

	query := `
		INSERT INTO posts (content, title, user_id, tags)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
	`

	ctxWTimeout, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	err := p.db.QueryRowContext(
		ctxWTimeout,
		query,
		post.Content,
		post.Title,
		post.UserId,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (p PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {

	query := `
		SELECT id, user_id, title, content, created_at,  updated_at, tags, version
		FROM posts
		WHERE id = $1
	`

	ctxWTimeout, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	var post Post
	err := p.db.QueryRowContext(ctxWTimeout, query, id).Scan(
		&post.ID,
		&post.UserId,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		pq.Array(&post.Tags),
		&post.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (p PostStore) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE 
		FROM posts
		WHERE id = $1
	`

	ctxWTimeout, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	res, err := p.db.ExecContext(ctxWTimeout, query, id)

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

func (p PostStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1, content = $2, version = version + 1
		WHERE id = $3 and version = $4
		RETURNING version
	`
	ctxWTimeout, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	err := p.db.QueryRowContext(
		ctxWTimeout,
		query,
		post.Title,
		post.Content,
		post.ID,
		post.Version,
	).Scan(
		&post.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}
