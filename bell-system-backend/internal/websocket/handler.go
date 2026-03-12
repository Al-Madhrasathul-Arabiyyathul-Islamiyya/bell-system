package websocket

import (
	"net/http"
	"time"

	"arabiyya.edu.mv/bell-system-backend/config"
	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	ws "github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// Handler handles WebSocket upgrade requests.
type Handler struct {
	Hub    *Hub
	Tokens handlers.TokenService
	Logger *logger.Logger
	Config config.WebSocketConfig
}

// NewHandler creates a new WebSocket Handler.
func NewHandler(hub *Hub, tokens handlers.TokenService, log *logger.Logger, cfg config.WebSocketConfig) *Handler {
	return &Handler{
		Hub:    hub,
		Tokens: tokens,
		Logger: log,
		Config: cfg,
	}
}

// ServeHTTP upgrades the HTTP connection to WebSocket after JWT authentication.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, `{"error":{"code":"missing_token","message":"token query parameter is required"}}`, http.StatusUnauthorized)
		return
	}

	claims, err := h.Tokens.ValidateToken(token)
	if err != nil {
		http.Error(w, `{"error":{"code":"invalid_token","message":"invalid or expired token"}}`, http.StatusUnauthorized)
		return
	}

	clientType := r.URL.Query().Get("client_type")
	if clientType != "admin" && clientType != "client" {
		http.Error(w, `{"error":{"code":"invalid_client_type","message":"client_type must be admin or client"}}`, http.StatusBadRequest)
		return
	}

	if clientType == "admin" && claims.Role != "admin" {
		http.Error(w, `{"error":{"code":"forbidden","message":"admin client_type requires admin role"}}`, http.StatusForbidden)
		return
	}

	clientName := r.URL.Query().Get("client_name")

	conn, err := ws.Accept(w, r, &ws.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		h.Logger.Error("websocket accept failed", err)
		return
	}

	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Real-IP"); forwarded != "" {
		ip = forwarded
	}

	client := NewClient(conn, h.Hub, h.Logger, h.Config, ip, clientType, clientName, claims.UserID)
	h.Hub.RegisterClient(client)

	ack := models.WebSocketMessage{
		Type:      "connection_acknowledged",
		Timestamp: time.Now().UTC(),
		Payload: models.ConnectionAcknowledgedPayload{
			ConnectionID: client.ID,
			ClientType:   clientType,
			Message:      "Connection successful",
		},
	}
	_ = wsjson.Write(r.Context(), conn, ack)

	go client.WritePump(r.Context())
	client.ReadPump(r.Context())
}
