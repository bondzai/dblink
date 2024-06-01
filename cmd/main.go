package main

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/bondzai/dblink/internal/interfaces/handlers"
	"github.com/bondzai/dblink/internal/interfaces/repositories"
	"github.com/bondzai/dblink/internal/services"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var ctx = context.Background()

func init() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	mdb := repositories.NewMockRepo()

	users := mdb.MockUsers()

	for _, user := range users {
		data, err := json.Marshal(user)
		if err != nil {
			log.Fatalf("Could not marshal user: %v", err)
		}

		err = rdb.Set(ctx, strconv.Itoa(user.UserID), data, 0).Err()
		if err != nil {
			log.Fatalf("Could not set user data in Redis: %v", err)
		}
	}

	log.Println("Successfully seeded Redis with initial user data.")
}

func main() {
	app := fiber.New()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	redisRepo := repositories.NewRedisRepository(rdb)
	userService := services.NewUserService(redisRepo)
	wsHandler := handlers.NewWebSocketHandler(userService)

	app.Use("/ws", wsHandler.HandleWsUpgrade)

	app.Get("/ws/read/:id", websocket.New(wsHandler.ReadUser))
	app.Get("/ws/update/:id", websocket.New(wsHandler.UpdateUser))

	app.Listen(":8080")
}
