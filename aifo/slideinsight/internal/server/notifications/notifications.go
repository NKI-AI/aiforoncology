// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package notifications

import (
	"context"
	"fmt"
	"sync"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2/log"
)

// Hub manages WebSocket connections for notifications
type Hub struct {
	service     services.NotificationService
	connections map[int][]*websocket.Conn // userID -> connections
	mutex       sync.RWMutex
}

// NewHub creates a new notification hub
func NewHub(service services.NotificationService) *Hub {
	return &Hub{
		service:     service,
		connections: make(map[int][]*websocket.Conn),
	}
}

// RegisterConnection registers a WebSocket connection for a user
func (h *Hub) RegisterConnection(userID int, conn *websocket.Conn) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.connections[userID] = append(h.connections[userID], conn)
	h.service.RegisterWebSocketConnection(userID, conn)

	log.Info(fmt.Sprintf("WebSocket connection registered: userID=%d, total=%d", userID, len(h.connections[userID])))
}

// UnregisterConnection removes a WebSocket connection for a user
func (h *Hub) UnregisterConnection(userID int, conn *websocket.Conn) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	connections := h.connections[userID]
	for i, existingConn := range connections {
		if existingConn == conn {
			h.connections[userID] = append(connections[:i], connections[i+1:]...)
			break
		}
	}

	if len(h.connections[userID]) == 0 {
		delete(h.connections, userID)
	}

	h.service.UnregisterWebSocketConnection(userID, conn)

	log.Info(fmt.Sprintf("WebSocket connection unregistered: userID=%d", userID))
}

// BroadcastToUser sends a message to all connections for a user
func (h *Hub) BroadcastToUser(userID int, message domain.WebSocketMessage) {
	h.service.BroadcastToUser(userID, message)
}

// GetNotificationStats gets notification statistics for a user
func (h *Hub) GetNotificationStats(ctx context.Context, userID int) (*domain.NotificationStats, error) {
	return h.service.GetNotificationStats(ctx, userID)
}

// Close closes all connections
func (h *Hub) Close() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for userID, connections := range h.connections {
		for _, conn := range connections {
			conn.Close()
		}
		delete(h.connections, userID)
	}

	h.service.Close()
}
