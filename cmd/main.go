package main

import (
	"sync"

	"github.com/bondzai/dblink/internal/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

var (
	drivers  = make(map[string]domain.DriverDTO)
	clients  = make(map[string]map[*websocket.Conn]bool)
	muRead   = &sync.Mutex{}
	muUpdate = &sync.Mutex{}
)

func main() {
	app := fiber.New()

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		driverID := c.Params("id")

		muRead.Lock()
		if clients[driverID] == nil {
			clients[driverID] = make(map[*websocket.Conn]bool)
		}
		clients[driverID][c] = true
		muRead.Unlock()

		dto := getLatestData(driverID)
		if err := c.WriteJSON(dto); err != nil {
			return
		}

		defer func() {
			muRead.Lock()
			delete(clients[driverID], c)
			if len(clients[driverID]) == 0 {
				delete(clients, driverID)
			}
			muRead.Unlock()
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

func getLatestData(driverID string) domain.DriverDTO {
	muRead.Lock()
	defer muRead.Unlock()

	if driver, exists := drivers[driverID]; exists {
		return driver
	}

	return domain.DriverDTO{
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
	muUpdate.Lock()
	defer muUpdate.Unlock()

	driver, exists := drivers[driverID]
	if !exists {
		driver = domain.DriverDTO{}
	}

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

	drivers[driverID] = driver
	broadcastLocation(driverID, driver)
}

func broadcastLocation(driverID string, driver domain.DriverDTO) {
	muRead.Lock()
	defer muRead.Unlock()

	for client := range clients[driverID] {
		if err := client.WriteJSON(driver); err != nil {
			client.Close()
			delete(clients[driverID], client)
		}
	}
}
