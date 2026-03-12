package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	"go.uber.org/zap"
)

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastMessage
	done       chan struct{}
	mu         sync.RWMutex
	logger     *logger.Logger
}

type broadcastMessage struct {
	data      []byte
	adminOnly bool
}

// NewHub creates a new Hub instance.
func NewHub(log *logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan broadcastMessage, 256),
		done:       make(chan struct{}),
		logger:     log,
	}
}

// Run starts the hub's main event loop. It blocks until ctx is cancelled.
func (h *Hub) Run(done <-chan struct{}) {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			h.logger.Info("client registered",
				zap.String("client_id", client.ID),
				zap.String("client_type", client.ClientType),
			)
			h.broadcastConnectedClients()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.send)
			}
			h.mu.Unlock()
			h.logger.Info("client unregistered",
				zap.String("client_id", client.ID),
				zap.String("client_type", client.ClientType),
			)
			h.broadcastConnectedClients()

		case msg := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				if msg.adminOnly && client.ClientType != "admin" {
					continue
				}
				client.SendMessage(msg.data)
			}
			h.mu.RUnlock()

		case <-done:
			h.mu.Lock()
			for _, client := range h.clients {
				close(client.send)
			}
			h.clients = make(map[string]*Client)
			h.mu.Unlock()
			close(h.done)
			return
		}
	}
}

// Done returns a channel that is closed when the hub has shut down.
func (h *Hub) Done() <-chan struct{} {
	return h.done
}

// RegisterClient adds a client to the hub.
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient removes a client from the hub.
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(msg models.WebSocketMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("failed to marshal broadcast message", err)
		return
	}
	h.broadcast <- broadcastMessage{data: data, adminOnly: false}
}

// BroadcastToAdmins sends a message only to admin clients.
func (h *Hub) BroadcastToAdmins(msg models.WebSocketMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("failed to marshal admin broadcast message", err)
		return
	}
	h.broadcast <- broadcastMessage{data: data, adminOnly: true}
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// ConnectedClients returns info about all connected clients.
func (h *Hub) ConnectedClients() []models.ClientInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients := make([]models.ClientInfo, 0, len(h.clients))
	for _, c := range h.clients {
		clients = append(clients, c.Info())
	}
	return clients
}

// broadcastConnectedClients sends the connected_clients event to all admin clients.
func (h *Hub) broadcastConnectedClients() {
	clients := h.ConnectedClients()
	msg := models.WebSocketMessage{
		Type:      "connected_clients",
		Timestamp: time.Now().UTC(),
		Payload: models.ConnectedClientsPayload{
			Clients: clients,
		},
	}
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("failed to marshal connected_clients message", err)
		return
	}
	for _, client := range h.clients {
		if client.ClientType != "admin" {
			continue
		}
		client.SendMessage(data)
	}
}
