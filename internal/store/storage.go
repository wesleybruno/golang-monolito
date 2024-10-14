package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrDuplicateKey = errors.New("duplicate key value violates unique constraint")
	TimeOutTime     = time.Second * 5
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, int64) (*Post, error)
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64, PaginationFeedQuery) ([]*PostWithMetadata, error)
	}
	Users interface {
		Create(context.Context, *User) error
		GetById(ctx context.Context, id int64) (*User, error)
	}
	Comments interface {
		GetByPostId(ctx context.Context, postId int64) ([]Comment, error)
		Create(context.Context, *Comment) error
	}
	Follower interface {
		Follow(ctx context.Context, currentId int64, followId int64) error
		Unfollow(ctx context.Context, currentId int64, followId int64) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:    &PostStore{db},
		Users:    &UserStore{db},
		Comments: &CommentStore{db},
		Follower: &FollowerStore{db},
	}
}
