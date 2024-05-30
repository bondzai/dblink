package main

import (
	"github.com/bondzai/dblink/internal/handlers"
	"github.com/bondzai/dblink/internal/repositories"
	"github.com/bondzai/dblink/internal/services"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	app := fiber.New()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	redisRepo := repositories.NewRedisRepository(rdb)
	userService := services.NewUserService(redisRepo)
	wsHandler := handlers.NewWebSocketHandler(userService)

	app.Get("/ws/read/:id", websocket.New(wsHandler.ReadUser))
	app.Get("/ws/update/:id", websocket.New(wsHandler.UpdateUser))

	app.Listen(":8080")
}
