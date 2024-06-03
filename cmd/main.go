package main

import (
	"context"

	"github.com/bondzai/dblink/internal/handler"
	"github.com/bondzai/dblink/internal/repository"
	"github.com/bondzai/dblink/internal/service"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	ctx := context.Background()

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	redisRepo := repository.NewRedisRepository(redisClient, ctx)
	driverService := service.NewDriverService(redisRepo)
	driverHandler := handler.NewDriverHandler(driverService)

	app := fiber.New()

	app.Get("/ws/:id", websocket.New(driverHandler.WebSocketHandler))

	app.Listen(":8080")
}
