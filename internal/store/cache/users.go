package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wesleybruno/golang-monolito/internal/store"
)

type UsersStore struct {
	rdb *redis.Client
}

var cacheExpTime = time.Minute * 3

func (s UsersStore) Get(ctx context.Context, id int64) (*store.User, error) {

	cacheKey := fmt.Sprintf("user-%v", id)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user store.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s UsersStore) Set(ctx context.Context, user *store.User) error {

	cacheKey := fmt.Sprintf("user-%v", user.ID)

	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	err = s.rdb.SetEx(ctx, cacheKey, json, cacheExpTime).Err()
	return err

}
