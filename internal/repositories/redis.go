package repositories

import (
	"context"
	"encoding/json"

	"github.com/bondzai/dblink/internal/models"

	"github.com/go-redis/redis/v8"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

func (r *RedisRepository) GetUser(ctx context.Context, id string) (*models.User, error) {
	data, err := r.client.Get(ctx, id).Result()
	if err != nil {
		return nil, err
	}
	var user models.User
	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *RedisRepository) UpdateUser(ctx context.Context, id string, user *models.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, id, data, 0).Err()
}
