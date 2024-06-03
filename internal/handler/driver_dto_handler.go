package handler

import (
	"sync"

	"github.com/bondzai/dblink/internal/domain"
	"github.com/bondzai/dblink/internal/service"
	"github.com/gofiber/websocket/v2"
)

type DriverHandler struct {
	service *service.DriverService
	clients map[string]map[*websocket.Conn]bool
	mu      *sync.Mutex
}

func NewDriverHandler(service *service.DriverService) *DriverHandler {
	return &DriverHandler{
		service: service,
		clients: make(map[string]map[*websocket.Conn]bool),
		mu:      &sync.Mutex{},
	}
}

func (h *DriverHandler) WebSocketHandler(c *websocket.Conn) {
	driverID := c.Params("id")

	h.mu.Lock()
	if h.clients[driverID] == nil {
		h.clients[driverID] = make(map[*websocket.Conn]bool)
	}
	h.clients[driverID][c] = true
	h.mu.Unlock()

	dto := h.service.GetLatestData(driverID)
	if err := c.WriteJSON(dto); err != nil {
		return
	}

	defer func() {
		h.mu.Lock()
		delete(h.clients[driverID], c)
		if len(h.clients[driverID]) == 0 {
			delete(h.clients, driverID)
		}
		h.mu.Unlock()
		c.Close()
	}()

	for {
		var updateData map[string]interface{}
		if err := c.ReadJSON(&updateData); err != nil {
			return
		}

		dto := h.service.ProcessUpdate(driverID, updateData)
		h.broadcastLocation(driverID, *dto)
		dto = nil
	}
}

func (h *DriverHandler) broadcastLocation(driverID string, driver domain.DriverDTO) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients[driverID] {
		if err := client.WriteJSON(driver); err != nil {
			client.Close()
			delete(h.clients[driverID], client)
		}
	}
}
