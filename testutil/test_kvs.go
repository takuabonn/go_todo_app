package testutil

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest"
)

func CreateRedisContainer() (*dockertest.Resource, *dockertest.Pool) {
	// Dockerとの接続
	pool, err := dockertest.NewPool("")

	pool.MaxWait = time.Minute * 3
	if err != nil {
		log.Fatalf("Could not connect to docker ee: %s", err)
	}

	// Redisコンテナの設定
	redisOpts := dockertest.RunOptions{
		Repository: "redis",
		Tag:        "latest",
	}

	// Redisコンテナの起動
	redisResource, err := pool.RunWithOptions(&redisOpts)
	if err != nil {
		log.Fatalf("Could not start Redis container: %s", err)
	}

	return redisResource, pool
}

func CloseRedisContainer(resource *dockertest.Resource, pool *dockertest.Pool) {
	// コンテナの終了
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func ConnectRedis(resource *dockertest.Resource, pool *dockertest.Pool) (*redis.Client, string) {
	// Redisコンテナが起動するまで待機
	var client *redis.Client
	port := resource.GetPort("6379/tcp")
	if err := pool.Retry(func() error {
		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("127.0.0.1:%s", port),
			Password: "",
			DB:       0, // default database number
		})
		return client.Ping(context.Background()).Err()
	}); err != nil {
		log.Fatalf("Could not connect to Redis: %s", err)
	}
	return client, port
}
