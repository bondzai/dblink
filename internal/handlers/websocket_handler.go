package handlers

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/bondzai/dblink/internal/models"
	"github.com/bondzai/dblink/internal/services"

	"github.com/gofiber/websocket/v2"
)

type WebSocketHandler struct {
	userService *services.UserService
}

func NewWebSocketHandler(userService *services.UserService) *WebSocketHandler {
	return &WebSocketHandler{
		userService: userService,
	}
}

func (h *WebSocketHandler) ReadUser(c *websocket.Conn) {
	defer c.Close()

	id := strings.TrimPrefix(c.Params("id"), "/ws/read/")
	user, err := h.userService.GetUser(context.Background(), id)
	if err != nil {
		c.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}

	data, err := json.Marshal(user)
	if err != nil {
		c.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}

	c.WriteMessage(websocket.TextMessage, data)
}

func (h *WebSocketHandler) UpdateUser(c *websocket.Conn) {
	defer c.Close()

	id := strings.TrimPrefix(c.Params("id"), "/ws/update/")
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			break
		}

		var location models.Location
		err = json.Unmarshal(message, &location)
		if err != nil {
			c.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			continue
		}

		user, err := h.userService.UpdateUser(context.Background(), id, &location)
		if err != nil {
			c.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			continue
		}

		data, err := json.Marshal(user)
		if err != nil {
			c.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			continue
		}

		c.WriteMessage(websocket.TextMessage, data)
	}
}
