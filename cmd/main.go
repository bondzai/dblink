package main

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type DriverDTO struct {
	Location     DriverLocation     `json:"location"`
	LoginSession DriverLoginSession `json:"loginSession"`
	Type         DriverType         `json:"type"`
	Job          *uuid.UUID         `json:"job"`
}

type DriverLocation struct {
	Lat  string `json:"lat"`
	Long string `json:"long"`
}

type DriverLoginSession struct {
	DeviceId string `json:"deviceId"`
}

type DriverType struct {
	CompanyApproveStatus int       `json:"companyApproveStatus"`
	JobAcceptStatus      *int      `json:"jobAcceptStatus"`
	IsInternalCompany    bool      `json:"isInternalCompany"`
	VehicleTypeID        uuid.UUID `json:"vehicleTypeId"`
}

var (
	drivers = make(map[string]DriverDTO)
	clients = make(map[string]map[*websocket.Conn]bool)
	lock    = &sync.Mutex{}
	mutex   = &sync.Mutex{}
)

func main() {
	app := fiber.New()

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		driverID := c.Params("id")

		lock.Lock()
		if clients[driverID] == nil {
			clients[driverID] = make(map[*websocket.Conn]bool)
		}
		clients[driverID][c] = true
		lock.Unlock()

		defer func() {
			lock.Lock()
			delete(clients[driverID], c)
			if len(clients[driverID]) == 0 {
				delete(clients, driverID)
			}
			lock.Unlock()
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

func processUpdate(driverID string, updateData map[string]interface{}) {
	mutex.Lock()
	defer mutex.Unlock()

	driver, exists := drivers[driverID]
	if !exists {
		driver = DriverDTO{}
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
			driver.LoginSession.DeviceId = deviceId
		}
	}

	drivers[driverID] = driver
	broadcastLocation(driverID, driver)
}

func broadcastLocation(driverID string, driver DriverDTO) {
	lock.Lock()
	defer lock.Unlock()

	for client := range clients[driverID] {
		if err := client.WriteJSON(driver); err != nil {
			client.Close()
			delete(clients[driverID], client)
		}
	}
}
