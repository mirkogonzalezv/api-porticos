package cache

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedis(host string, port int, password string, db int, useTLS bool) (*RedisClient, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	opts := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}
	if useTLS {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{Client: client}, nil
}
