package store

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/takuabonn/go_todo_app/config"
	"github.com/takuabonn/go_todo_app/entity"
)

func NewKVS(ctx context.Context, cfg *config.Config) (*KVS, error) {
	cli := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
	})
	if err := cli.Ping().Err(); err != nil {
		return nil, err
	}
	return &KVS{Cli: cli}, nil
}

type KVS struct {
	Cli *redis.Client
}

func (k *KVS) Save(ctx context.Context, key string, userID entity.UserID) error {
	id := int64(userID)
	return k.Cli.Set(key, id, 30*time.Minute).Err()
}

func (k *KVS) Load(ctx context.Context, key string) (entity.UserID, error) {
	id, err := k.Cli.Get(key).Int64()
	if err != nil {
		return 0, fmt.Errorf("failed to get by %q: %w", key, ErrNotFound)
	}
	return entity.UserID(id), nil
}
