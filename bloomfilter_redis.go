package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisBloomFilter struct {
	client *redis.Client
	key    string
}

func NewRedisBloomFilter(addr, key string, errorRate float64, capacity int64) (*RedisBloomFilter, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()

	err := rdb.Do(ctx, "BF.RESERVE", key, fmt.Sprintf("%f", errorRate), capacity).Err()
	if err != nil && err.Error() != "ERR item exists" {
		return nil, fmt.Errorf("failed to create bloom filter: %v", err)
	}

	return &RedisBloomFilter{
		client: rdb,
		key:    key,
	}, nil
}

func (rbf *RedisBloomFilter) Add(item string) (bool, error) {
	ctx := context.Background()
	added, err := rbf.client.Do(ctx, "BF.ADD", rbf.key, item).Bool()
	if err != nil {
		return false, fmt.Errorf("failed to add item: %v", err)
	}
	return added, nil
}

func (rbf *RedisBloomFilter) Check(item string) (bool, error) {
	ctx := context.Background()
	exists, err := rbf.client.Do(ctx, "BF.EXISTS", rbf.key, item).Bool()
	if err != nil {
		return false, fmt.Errorf("failed to check item: %v", err)
	}
	return exists, nil
}
