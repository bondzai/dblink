package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/bondzai/dblink/internal/domain"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

var (
	clients     = make(map[string]map[*websocket.Conn]bool)
	ctx         = context.Background()
	redisClient *redis.Client
	mu          = &sync.Mutex{}
)

const redisKeyPrefix = "driver:"

func main() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	app := fiber.New()

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		driverID := c.Params("id")

		mu.Lock()
		if clients[driverID] == nil {
			clients[driverID] = make(map[*websocket.Conn]bool)
		}
		clients[driverID][c] = true
		mu.Unlock()

		dto := getLatestData(driverID)
		if err := c.WriteJSON(dto); err != nil {
			return
		}

		defer func() {
			mu.Lock()
			delete(clients[driverID], c)
			if len(clients[driverID]) == 0 {
				delete(clients, driverID)
			}
			mu.Unlock()
			c.Close()
		}()

		for {
			var updateData map[string]interface{}
			if err := c.ReadJSON(&updateData); err != nil {
				return
			}

			processUpdate(driverID, updateData)
		}
	}))

	app.Listen(":8080")
}

// cache aside logic: get cache or db
func getLatestData(driverID string) domain.DriverDTO {
	key := redisKeyPrefix + driverID
	val, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		defaultDriver := getDefaultData(driverID)

		data, err := json.Marshal(defaultDriver)
		if err != nil {
			log.Printf("Error marshalling default driver data: %v", err)
			return domain.DriverDTO{}
		}

		if err := redisClient.Set(ctx, key, data, 0).Err(); err != nil {
			log.Printf("Error setting default driver data in Redis: %v", err)
			return domain.DriverDTO{}
		}

		return defaultDriver
	} else if err != nil {
		log.Printf("Error getting data from Redis: %v", err)
		return domain.DriverDTO{}
	}

	var driver domain.DriverDTO
	if err := json.Unmarshal([]byte(val), &driver); err != nil {
		log.Printf("Error unmarshalling data from Redis: %v", err)
		return domain.DriverDTO{}
	}

	return driver
}

// todo: query data from db
func getDefaultData(driverID string) domain.DriverDTO {
	return domain.DriverDTO{
		Id: driverID,
		Location: domain.DriverLocation{
			Lat:  "0",
			Long: "0",
		},
		LoginSession: domain.DriverLoginSession{
			DeviceID: "default-device-id",
		},
		Type: domain.DriverType{
			CompanyApproveStatus: 0,
			JobAcceptStatus:      nil,
			IsInternalCompany:    false,
			VehicleTypeID:        uuid.Nil,
		},
		Job: nil,
	}
}

func processUpdate(driverID string, updateData map[string]interface{}) {
	driver := getLatestData(driverID)

	if loc, ok := updateData["location"].(map[string]interface{}); ok {
		if lat, ok := loc["lat"].(string); ok {
			driver.Location.Lat = lat
		}
		if long, ok := loc["long"].(string); ok {
			driver.Location.Long = long
		}
	}

	if session, ok := updateData["loginSession"].(map[string]interface{}); ok {
		if deviceId, ok := session["deviceId"].(string); ok {
			driver.LoginSession.DeviceID = deviceId
		}
	}

	if driverType, ok := updateData["type"].(map[string]interface{}); ok {
		if isInternalCompany, ok := driverType["isInternalCompany"].(bool); ok {
			driver.Type.IsInternalCompany = isInternalCompany
		}
	}

	data, err := json.Marshal(driver)
	if err != nil {
		log.Printf("Error marshalling data to Redis: %v", err)
		return
	}

	key := redisKeyPrefix + driverID
	if err := redisClient.Set(ctx, key, data, 0).Err(); err != nil {
		log.Printf("Error setting data in Redis: %v", err)
		return
	}

	broadcastLocation(driverID, driver)
}

func broadcastLocation(driverID string, driver domain.DriverDTO) {
	mu.Lock()
	defer mu.Unlock()

	for client := range clients[driverID] {
		if err := client.WriteJSON(driver); err != nil {
			client.Close()
			delete(clients[driverID], client)
		}
	}
}
