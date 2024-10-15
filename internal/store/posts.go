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
	Comment   []Comment `json:"comments,omitempty"`
	Version   string    `json:"version"`
}

type PostWithMetadata struct {
	Post          Post  `json:"post"`
	User          User  `json:"user"`
	CountComments int64 `json:"total_comments"`
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

func (p PostStore) GetUserFeed(ctx context.Context, id int64, pagination PaginationFeedQuery) ([]*PostWithMetadata, error) {

	whereTags := ""
	if len(pagination.Tags) > 0 {
		whereTags = "and (p.tags @> $5)"
	}

	query := `
	SELECT 
			p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags,
			u.username,
			COUNT(c.id) AS comments_count
		FROM posts p
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN users u ON p.user_id = u.id
		JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
		WHERE 
			1=1
			AND f.user_id = $1 
			AND (p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%')
			` + whereTags + `
		GROUP BY p.id, u.username
		ORDER BY p.created_at ` + pagination.Sort + `
		LIMIT $2 OFFSET $3
`

	ctxWTimeout, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	arr := pq.Array(pagination.Tags)
	var rows *sql.Rows
	var err error

	if len(pagination.Tags) > 0 {
		rows, err = p.db.QueryContext(ctxWTimeout, query, id, pagination.Limit, pagination.Offset, pagination.Search, arr)
	} else {
		rows, err = p.db.QueryContext(ctxWTimeout, query, id, pagination.Limit, pagination.Offset, pagination.Search)

	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feed []*PostWithMetadata
	for rows.Next() {

		var p PostWithMetadata
		err := rows.Scan(
			&p.Post.ID,
			&p.Post.UserId,
			&p.Post.Title,
			&p.Post.Content,
			&p.Post.CreatedAt,
			&p.Post.Version,
			pq.Array(&p.Post.Tags),
			&p.User.Username,
			&p.CountComments,
		)
		if err != nil {
			return nil, err
		}

		feed = append(feed, &p)

	}

	return feed, nil
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
