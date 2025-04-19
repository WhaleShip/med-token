package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type refreshRepo struct {
	rdb *redis.Client
}

func NewRefreshRepo(rdb *redis.Client) *refreshRepo {
	return &refreshRepo{rdb: rdb}
}

func (r *refreshRepo) Save(ctx context.Context, jti, hash, ip, userID string, ttl time.Duration) error {
	key := "refresh:" + jti
	if err := r.rdb.HSet(ctx, key, map[string]interface{}{
		"hash":    hash,
		"ip":      ip,
		"user_id": userID,
	}).Err(); err != nil {
		return err
	}
	return r.rdb.Expire(ctx, key, ttl).Err()
}

func (r *refreshRepo) Get(ctx context.Context, jti string) (string, string, string, error) {
	key := "refresh:" + jti
	vals, err := r.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return "", "", "", err
	}
	if len(vals) == 0 {
		return "", "", "", errors.New("not found")
	}
	return vals["hash"], vals["ip"], vals["user_id"], nil
}

func (r *refreshRepo) Delete(ctx context.Context, jti string) error {
	return r.rdb.Del(ctx, "refresh:"+jti).Err()
}
