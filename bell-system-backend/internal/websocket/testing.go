package websocket

import (
	"time"

	"github.com/google/uuid"
)

// NewTestClient creates a Client without a real WebSocket connection, for testing.
// The send channel is accessible via ReadSend and DrainSend methods.
func NewTestClient(hub *Hub, clientType string, clientName string) *Client {
	return &Client{
		ID:             uuid.New().String(),
		IP:             "127.0.0.1",
		ClientType:     clientType,
		ClientName:     clientName,
		ConnectedSince: time.Now().UTC(),
		UserID:         uuid.New(),
		hub:            hub,
		send:           make(chan []byte, sendBufferSize),
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
