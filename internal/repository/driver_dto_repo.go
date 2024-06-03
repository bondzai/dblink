package repository

import (
	"context"
	"encoding/json"

	"github.com/bondzai/dblink/internal/domain"
	"github.com/go-redis/redis/v8"
)

type RedisRepository struct {
	client *redis.Client
	ctx    context.Context
}

const redisKeyPrefix = "driver:"

func NewRedisRepository(client *redis.Client, ctx context.Context) *RedisRepository {
	return &RedisRepository{
		client: client,
		ctx:    ctx,
	}
}

func (r *RedisRepository) GetDriver(driverID string) (*domain.DriverDTO, error) {
	key := redisKeyPrefix + driverID

	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return &domain.DriverDTO{}, nil
	} else if err != nil {
		return &domain.DriverDTO{}, err
	}

	var driver domain.DriverDTO
	if err := json.Unmarshal([]byte(val), &driver); err != nil {
		return &domain.DriverDTO{}, err
	}

	return &driver, nil
}

func (r *RedisRepository) SaveDriver(driver *domain.DriverDTO) error {
	key := redisKeyPrefix + driver.Id
	data, err := json.Marshal(driver)
	if err != nil {
		return err
	}

	return r.client.Set(r.ctx, key, data, 0).Err()
}
