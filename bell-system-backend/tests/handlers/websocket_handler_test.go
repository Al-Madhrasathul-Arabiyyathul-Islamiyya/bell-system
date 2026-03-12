package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/config"
	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"
	ws "arabiyya.edu.mv/bell-system-backend/internal/websocket"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"
	"arabiyya.edu.mv/bell-system-backend/tests/mocks"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newWSTestServer(t *testing.T, tokenSvc handlers.TokenService) *httptest.Server {
	t.Helper()
	log, err := logger.New("test")
	require.NoError(t, err)

	cfg := config.WebSocketConfig{
		PingInterval:   30,
		PongTimeout:    10,
		MaxMessageSize: 512,
	}

	hub := ws.NewHub(log)
	done := make(chan struct{})
	go hub.Run(done)
	t.Cleanup(func() {
		close(done)
		<-hub.Done()
	})

	handler := ws.NewHandler(hub, tokenSvc, log, cfg)
	return httptest.NewServer(handler)
}

func validTokenService() *mocks.MockTokenService {
	return &mocks.MockTokenService{
		ValidateTokenFunc: func(token string) (*handlers.TokenClaims, error) {
			role := models.Role("admin")
			if token == "client-token" {
				role = "morning_user"
			}
			return &handlers.TokenClaims{
				UserID:   uuid.New(),
				Username: "testuser",
				Role:     role,
			}, nil
		},
	}
}

func TestWSHandler_MissingToken(t *testing.T) {
	srv := newWSTestServer(t, validTokenService())
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, srv.URL+"?client_type=client", nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWSHandler_InvalidToken(t *testing.T) {
	tokenSvc := &mocks.MockTokenService{
		ValidateTokenFunc: func(token string) (*handlers.TokenClaims, error) {
			return nil, assert.AnError
		},
	}
	srv := newWSTestServer(t, tokenSvc)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, srv.URL+"?token=bad&client_type=client", nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWSHandler_InvalidClientType(t *testing.T) {
	srv := newWSTestServer(t, validTokenService())
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, srv.URL+"?token=valid&client_type=invalid", nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestWSHandler_AdminClientTypeRequiresAdminRole(t *testing.T) {
	srv := newWSTestServer(t, validTokenService())
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// client-token maps to morning_user role
	_, resp, err := websocket.Dial(ctx, srv.URL+"?token=client-token&client_type=admin", nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestWSHandler_SuccessfulConnection(t *testing.T) {
	srv := newWSTestServer(t, validTokenService())
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, srv.URL+"?token=valid&client_type=admin&client_name=TestAdmin", nil)
	require.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Should receive connection_acknowledged
	var msg models.WebSocketMessage
	err = wsjson.Read(ctx, conn, &msg)
	require.NoError(t, err)
	assert.Equal(t, "connection_acknowledged", msg.Type)

	payloadBytes, _ := json.Marshal(msg.Payload)
	var ack models.ConnectionAcknowledgedPayload
	require.NoError(t, json.Unmarshal(payloadBytes, &ack))
	assert.Equal(t, "admin", ack.ClientType)
	assert.Equal(t, "Connection successful", ack.Message)
	assert.NotEmpty(t, ack.ConnectionID)
}

func TestWSHandler_ClientConnection(t *testing.T) {
	srv := newWSTestServer(t, validTokenService())
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, srv.URL+"?token=client-token&client_type=client&client_name=Display1", nil)
	require.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	var msg models.WebSocketMessage
	err = wsjson.Read(ctx, conn, &msg)
	require.NoError(t, err)
	assert.Equal(t, "connection_acknowledged", msg.Type)
}
