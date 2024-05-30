package handlers

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/bondzai/dblink/internal/models"
	"github.com/bondzai/dblink/internal/services"
	"github.com/gofiber/websocket/v2"
)

type WebSocketHandler struct {
	userService *services.UserService
	connections map[string]map[*websocket.Conn]struct{}
	mutex       sync.Mutex
}

func NewWebSocketHandler(userService *services.UserService) *WebSocketHandler {
	return &WebSocketHandler{
		userService: userService,
		connections: make(map[string]map[*websocket.Conn]struct{}),
	}
}

func (h *WebSocketHandler) ReadUser(c *websocket.Conn) {
	defer func() {
		h.mutex.Lock()
		for id := range h.connections {
			delete(h.connections[id], c)
		}
		h.mutex.Unlock()
	}()

	id := c.Params("id")

	user, err := h.userService.GetUser(context.Background(), id)
	if err != nil {
		c.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		return
	}

	data, err := json.Marshal(user)
	if err != nil {
		c.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		return
	}

	c.WriteMessage(websocket.TextMessage, data)

	h.mutex.Lock()
	if _, ok := h.connections[id]; !ok {
		h.connections[id] = make(map[*websocket.Conn]struct{})
	}
	h.connections[id][c] = struct{}{}
	h.mutex.Unlock()

	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			log.Println("ReadMessage error:", err)
			break
		}
	}
}

func (h *WebSocketHandler) UpdateUser(c *websocket.Conn) {
	defer c.Close()

	id := c.Params("id")
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("ReadMessage error:", err)
			break
		}

		var location models.Location
		if err := json.Unmarshal(message, &location); err != nil {
			c.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
			continue
		}

		user, err := h.userService.UpdateUser(context.Background(), id, &location)
		if err != nil {
			c.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
			continue
		}

		data, err := json.Marshal(user)
		if err != nil {
			c.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
			continue
		}

		h.mutex.Lock()
		for conn := range h.connections[id] {
			conn.WriteMessage(websocket.TextMessage, data)
		}
		h.mutex.Unlock()
	}
}
