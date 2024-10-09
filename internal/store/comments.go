package store

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID        int64  `json:"id"`
	UserId    int64  `json:"user_id"`
	PostId    int64  `json:"post_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      User   `json:"user"`
}

type CommentStore struct {
	db *sql.DB
}

func (c CommentStore) GetByPostId(ctx context.Context, postId int64) ([]Comment, error) {

	query := `
		SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, users.username, users.id  FROM comments c
		JOIN users on users.id = c.user_id
		WHERE c.post_id = $1
		ORDER BY c.created_at DESC;
	`

	ctxWTimeout, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	rows, err := c.db.QueryContext(ctxWTimeout, query, postId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		c.User = User{}
		err := rows.Scan(&c.ID, &c.PostId, &c.UserId, &c.Content, &c.CreatedAt, &c.User.Username, &c.User.ID)
		if err != nil {
			return nil, err
		}

		comments = append(comments, c)

	}

	return comments, nil
}

func (c CommentStore) Create(ctx context.Context, comment *Comment) error {

	query := `
		INSERT INTO comments (user_id, post_id, content)
		VALUES ($1, $2, $3) RETURNING id, created_at
	`

	ctxWTimeout, cancel := context.WithTimeout(ctx, TimeOutTime)
	defer cancel()

	err := c.db.QueryRowContext(
		ctxWTimeout,
		query,
		comment.UserId,
		comment.PostId,
		comment.Content,
	).Scan(
		&comment.ID,
		&comment.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}
