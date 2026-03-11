package services

import (
	"fmt"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var _ handlers.TokenService = (*TokenService)(nil)

type TokenService struct {
	secret    string
	expiresIn int // minutes
}

func NewTokenService(secret string, expiresIn int) *TokenService {
	return &TokenService{secret: secret, expiresIn: expiresIn}
}

func (s *TokenService) GenerateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":      user.ID.String(),
		"username": user.Username,
		"role":     string(user.Role),
		"exp":      time.Now().Add(time.Duration(s.expiresIn) * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *TokenService) ValidateToken(tokenStr string) (*handlers.TokenClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return nil, fmt.Errorf("missing subject claim: %w", err)
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	username, _ := claims["username"].(string)
	role, _ := claims["role"].(string)

	return &handlers.TokenClaims{
		UserID:   userID,
		Username: username,
		Role:     models.Role(role),
	}, nil
}
