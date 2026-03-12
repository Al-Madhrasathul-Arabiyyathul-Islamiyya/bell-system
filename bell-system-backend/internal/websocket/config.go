package websocket

import "time"

// Config holds WebSocket server configuration.
type Config struct {
	PingInterval   time.Duration
	PongTimeout    time.Duration
	MaxMessageSize int64
}

// DefaultConfig returns sensible defaults for WebSocket configuration.
func DefaultConfig() Config {
	return Config{
		PingInterval:   30 * time.Second,
		PongTimeout:    10 * time.Second,
		MaxMessageSize: 512,
	}
}
