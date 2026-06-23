package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type RedisClient struct {
	*redis.Client
}

func New(addrs []string, password string, db, poolSize int) (*RedisClient, error) {
	cli := redis.NewClient(&redis.Options{
		Addr:         addrs[0],
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	if err := cli.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	log.Info().Msg("redis connected")
	return &RedisClient{cli}, nil
}

func (r *RedisClient) Close() error {
	return r.Client.Close()
}
