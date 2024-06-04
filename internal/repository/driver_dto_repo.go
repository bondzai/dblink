package repository

import (
	"context"
	"encoding/json"

	"github.com/bondzai/dblink/internal/domain"
	"github.com/go-redis/redis/v8"
)

type RedisRepository interface {
	GetDriver(driverID string) (*domain.DriverWsDto, error)
	SaveDriver(driver *domain.DriverWsDto) error
}

type redisRepository struct {
	client *redis.Client
	ctx    context.Context
}

const redisKeyPrefix = "driver:"

func NewRedisRepository(client *redis.Client, ctx context.Context) RedisRepository {
	return &redisRepository{
		client: client,
		ctx:    ctx,
	}
}

func (r *redisRepository) GetDriver(driverID string) (*domain.DriverWsDto, error) {
	key := redisKeyPrefix + driverID

	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return &domain.DriverWsDto{}, nil
	} else if err != nil {
		return &domain.DriverWsDto{}, err
	}

	var driver domain.DriverWsDto
	if err := json.Unmarshal([]byte(val), &driver); err != nil {
		return &domain.DriverWsDto{}, err
	}

	return &driver, nil
}

func (r *redisRepository) SaveDriver(driver *domain.DriverWsDto) error {
	key := redisKeyPrefix + driver.Id
	data, err := json.Marshal(driver)
	if err != nil {
		return err
	}

	return r.client.Set(r.ctx, key, data, 0).Err()
}
