package store

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/takuabonn/go_todo_app/config"
	"github.com/takuabonn/go_todo_app/entity"
)

func NewKVS(ctx context.Context, cfg *config.Config) (*KVS, error) {
	cli := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
	})
	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &KVS{Cli: cli}, nil
}

type KVS struct {
	Cli *redis.Client
}

func (k *KVS) Save(ctx context.Context, key string, userID entity.UserID) error {
	id := int64(userID)
	return k.Cli.Set(ctx, key, id, 30*time.Minute).Err()
}

func (k *KVS) Load(ctx context.Context, key string) (entity.UserID, error) {
	val, err := k.Cli.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, fmt.Errorf("failed to get by %q: %w", key, ErrNotFound)
	} else if err != nil {
		return 0, fmt.Errorf("failed to get by %q: %w", key, err)
	}
	id, _ := strconv.ParseInt(val, 10, 64)
	return entity.UserID(id), nil
}
