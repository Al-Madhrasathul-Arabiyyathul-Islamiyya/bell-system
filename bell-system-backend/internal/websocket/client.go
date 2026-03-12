package websocket

import (
	"context"
	"encoding/json"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const sendBufferSize = 256

// Client represents a single WebSocket connection.
type Client struct {
	ID             string
	IP             string
	ClientType     string
	ClientName     string
	ConnectedSince time.Time
	UserID         uuid.UUID

	conn   *websocket.Conn
	hub    *Hub
	send   chan []byte
	logger *logger.Logger
	config Config
}

// NewClient creates a new Client from an accepted WebSocket connection.
func NewClient(
	conn *websocket.Conn,
	hub *Hub,
	log *logger.Logger,
	cfg Config,
	ip string,
	clientType string,
	clientName string,
	userID uuid.UUID,
) *Client {
	return &Client{
		ID:             uuid.New().String(),
		IP:             ip,
		ClientType:     clientType,
		ClientName:     clientName,
		ConnectedSince: time.Now().UTC(),
		UserID:         userID,
		conn:           conn,
		hub:            hub,
		send:           make(chan []byte, sendBufferSize),
		logger:         log,
		config:         cfg,
	}
}

// Info returns the ClientInfo model for this client.
func (c *Client) Info() models.ClientInfo {
	return models.ClientInfo{
		ID:             c.ID,
		IP:             c.IP,
		ClientType:     c.ClientType,
		ClientName:     c.ClientName,
		ConnectedSince: c.ConnectedSince,
	}
}

// SendMessage queues a message for delivery. Non-blocking; drops if buffer full.
func (c *Client) SendMessage(data []byte) {
	select {
	case c.send <- data:
	default:
		c.logger.Info("dropping message, client send buffer full",
			zap.String("client_id", c.ID),
		)
	}
}

// ReadPump reads messages from the WebSocket connection.
// It runs until the connection is closed or an error occurs.
func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.hub.UnregisterClient(c)
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	conn := c.conn
	conn.SetReadLimit(c.config.MaxMessageSize)

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				c.logger.Info("client disconnected normally", zap.String("client_id", c.ID))
			} else {
				c.logger.Info("client read error",
					zap.String("client_id", c.ID),
					zap.Error(err),
				)
			}
			return
		}

		var msg models.WebSocketMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			c.logger.Info("invalid message from client",
				zap.String("client_id", c.ID),
				zap.Error(err),
			)
			continue
		}

		c.handleMessage(ctx, msg)
	}
}

// WritePump writes messages to the WebSocket connection.
// It runs until the send channel is closed.
func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(c.config.PingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		select {
		case data, ok := <-c.send:
			if !ok {
				return
			}
			err := c.conn.Write(ctx, websocket.MessageText, data)
			if err != nil {
				c.logger.Info("client write error",
					zap.String("client_id", c.ID),
					zap.Error(err),
				)
				return
			}

		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, c.config.PongTimeout)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				c.logger.Info("client ping failed",
					zap.String("client_id", c.ID),
					zap.Error(err),
				)
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// handleMessage processes an incoming client message.
func (c *Client) handleMessage(ctx context.Context, msg models.WebSocketMessage) {
	switch msg.Type {
	case "register":
		c.handleRegister(ctx, msg)
	case "heartbeat":
		// heartbeat acknowledged — no action needed, the read itself keeps the connection alive
	default:
		c.logger.Info("unknown message type from client",
			zap.String("client_id", c.ID),
			zap.String("type", msg.Type),
		)
	}
}

// handleRegister processes a registration message, updating client metadata.
func (c *Client) handleRegister(ctx context.Context, msg models.WebSocketMessage) {
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		return
	}
	var reg models.RegistrationPayload
	if err := json.Unmarshal(payloadBytes, &reg); err != nil {
		return
	}
	if reg.ClientName != "" {
		c.ClientName = reg.ClientName
	}
	c.logger.Info("client registered via message",
		zap.String("client_id", c.ID),
		zap.String("client_name", c.ClientName),
	)

	ack := models.WebSocketMessage{
		Type:      "connection_acknowledged",
		Timestamp: time.Now().UTC(),
		Payload: models.ConnectionAcknowledgedPayload{
			ConnectionID: c.ID,
			ClientType:   c.ClientType,
			Message:      "Registration updated",
		},
	}
	_ = wsjson.Write(ctx, c.conn, ack)
}
