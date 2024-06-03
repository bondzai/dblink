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
	DeviceID string `json:"deviceId"`
}

type DriverType struct {
	CompanyApproveStatus int       `json:"companyApproveStatus"`
	JobAcceptStatus      *int      `json:"jobAcceptStatus"`
	IsInternalCompany    bool      `json:"isInternalCompany"`
	VehicleTypeID        uuid.UUID `json:"vehicleTypeId"`
}

var (
	drivers  = make(map[string]DriverDTO)
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

func getLatestData(driverID string) DriverDTO {
	muRead.Lock()
	defer muRead.Unlock()

	if driver, exists := drivers[driverID]; exists {
		return driver
	}

	return DriverDTO{
		Location: DriverLocation{
			Lat:  "0",
			Long: "0",
		},
		LoginSession: DriverLoginSession{
			DeviceID: "default-device-id",
		},
		Type: DriverType{
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

func broadcastLocation(driverID string, driver DriverDTO) {
	muRead.Lock()
	defer muRead.Unlock()

	for client := range clients[driverID] {
		if err := client.WriteJSON(driver); err != nil {
			client.Close()
			delete(clients[driverID], client)
		}
	}
}
