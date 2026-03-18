package testutil

import (
	"net/http"

	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"

	"github.com/google/uuid"
)

// TestToken is the fixed bearer token accepted by PermissiveTokenService.
const TestToken = "test-valid-token"

// PermissiveTokenService returns valid admin claims for any token matching TestToken.
// Use this in handler tests where auth is not the focus.
var PermissiveTokenService handlers.TokenService = &permissiveTokenSvc{}

type permissiveTokenSvc struct{}

func (p *permissiveTokenSvc) GenerateToken(user *models.User) (string, error) {
	return TestToken, nil
}

func (p *permissiveTokenSvc) ValidateToken(token string) (*handlers.TokenClaims, error) {
	return &handlers.TokenClaims{
		UserID:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Username: "testadmin",
		Role:     models.RoleAdmin,
	}, nil
}

// SetAuthHeader adds a valid Bearer token to the request.
func SetAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+TestToken)
}
