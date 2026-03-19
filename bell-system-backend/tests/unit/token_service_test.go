package unit_test

import (
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenService(t *testing.T) {
	svc := services.NewTokenService("secret", 60)
	require.NotNil(t, svc)
}

func TestTokenService_GenerateToken(t *testing.T) {
	svc := services.NewTokenService("test-secret", 60)

	user := &models.User{
		ID:       uuid.New(),
		Username: "admin",
		Role:     models.RoleAdmin,
	}

	token, err := svc.GenerateToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestTokenService_ValidateToken_Success(t *testing.T) {
	svc := services.NewTokenService("test-secret", 60)
	userID := uuid.New()

	user := &models.User{
		ID:       userID,
		Username: "admin",
		Role:     models.RoleAdmin,
	}

	token, err := svc.GenerateToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, "admin", claims.Username)
	assert.Equal(t, models.RoleAdmin, claims.Role)
}

func TestTokenService_ValidateToken_Expired(t *testing.T) {
	svc := services.NewTokenService("test-secret", -1)

	user := &models.User{
		ID:       uuid.New(),
		Username: "admin",
		Role:     models.RoleAdmin,
	}

	token, err := svc.GenerateToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateToken(token)
	require.Error(t, err)
	assert.Nil(t, claims)
}

func TestTokenService_ValidateToken_WrongSecret(t *testing.T) {
	svcA := services.NewTokenService("secret-A", 60)
	svcB := services.NewTokenService("secret-B", 60)

	user := &models.User{
		ID:       uuid.New(),
		Username: "admin",
		Role:     models.RoleAdmin,
	}

	token, err := svcA.GenerateToken(user)
	require.NoError(t, err)

	claims, err := svcB.ValidateToken(token)
	require.Error(t, err)
	assert.Nil(t, claims)
}

func TestTokenService_ValidateToken_Malformed(t *testing.T) {
	svc := services.NewTokenService("test-secret", 60)

	claims, err := svc.ValidateToken("not.a.valid.token")
	require.Error(t, err)
	assert.Nil(t, claims)
}

func TestTokenService_ValidateToken_InvalidSigningMethod(t *testing.T) {
	svc := services.NewTokenService("test-secret", 60)

	// Create an RS256-signed token (not HMAC) — should be rejected
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub":      uuid.New().String(),
		"username": "admin",
		"role":     "admin",
		"exp":      time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	claims, err := svc.ValidateToken(tokenStr)
	require.Error(t, err)
	assert.Nil(t, claims)
}

func TestTokenService_ValidateToken_MissingSubject(t *testing.T) {
	secret := "test-secret"

	// Craft a token without a "sub" claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "admin",
		"role":     "admin",
		"exp":      time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	svc := services.NewTokenService(secret, 60)
	claims, err := svc.ValidateToken(tokenStr)
	require.Error(t, err)
	assert.Nil(t, claims)
}

func TestTokenService_ValidateToken_NonUUIDSubject(t *testing.T) {
	secret := "test-secret"

	// Manually craft a token with a non-UUID subject
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      "not-a-uuid",
		"username": "admin",
		"role":     "admin",
		"exp":      time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	svc := services.NewTokenService(secret, 60)
	claims, err := svc.ValidateToken(tokenStr)
	require.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "invalid user ID")
}

func TestTokenService_RoundTrip_AllRoles(t *testing.T) {
	svc := services.NewTokenService("test-secret", 60)

	roles := []models.Role{
		models.RoleAdmin,
		models.RoleMorningUser,
		models.RoleAfternoonUser,
	}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			userID := uuid.New()
			user := &models.User{
				ID:       userID,
				Username: string(role) + "_user",
				Role:     role,
			}

			token, err := svc.GenerateToken(user)
			require.NoError(t, err)

			claims, err := svc.ValidateToken(token)
			require.NoError(t, err)
			assert.Equal(t, userID, claims.UserID)
			assert.Equal(t, string(role)+"_user", claims.Username)
			assert.Equal(t, role, claims.Role)
		})
	}
}
