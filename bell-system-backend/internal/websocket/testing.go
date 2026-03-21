package websocket

import (
	"time"

	"arabiyya.edu.mv/bell-system-backend/config"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NewTestClient creates a Client without a real WebSocket connection, for testing.
// The send channel is accessible via ReadSend and DrainSend methods.
func NewTestClient(hub *Hub, clientType string, clientName string) *Client {
	nopLogger := &logger.Logger{Logger: zap.NewNop()}
	return &Client{
		ID:             uuid.New().String(),
		IP:             "127.0.0.1",
		ClientType:     clientType,
		clientName:     clientName,
		ConnectedSince: time.Now().UTC(),
		UserID:         uuid.New(),
		hub:            hub,
		send:           make(chan []byte, sendBufferSize),
		logger:         nopLogger,
	}
}

// ReadSend returns the next message from the send channel, or nil if empty.
func (c *Client) ReadSend() []byte {
	select {
	case data := <-c.send:
		return data
	default:
		return nil
	}
}

// DrainSend discards all pending messages in the send channel.
func (c *Client) DrainSend() {
	for {
		select {
		case <-c.send:
		default:
			return
		}
	}
}

// CloseConn closes the underlying WebSocket connection (for testing write-error paths).
func (c *Client) CloseConn() {
	c.conn.CloseNow()
}

// TestWSConfig returns a WebSocketConfig with short intervals suitable for tests.
func TestWSConfig() config.WebSocketConfig {
	return config.WebSocketConfig{
		PingInterval:   1,
		PongTimeout:    1,
		MaxMessageSize: 512,
	}
}
