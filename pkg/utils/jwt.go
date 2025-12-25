package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenPurpose string

const (
	PurposeManagement TokenPurpose = "management"
	PurposeBackoffice TokenPurpose = "backoffice"
	PurposeTarget     TokenPurpose = "target"
)

type Claims struct {
	UserID   string       `json:"user_id"`
	TenantID string       `json:"tenant_id"`
	Role     string       `json:"role,omitempty"`
	Groups   []string     `json:"groups,omitempty"` // Google Workspace groups
	Purpose  TokenPurpose `json:"purpose"`
	jwt.RegisteredClaims
}

func GenerateToken(secret string, purpose TokenPurpose, userID, tenantID, role string, groups []string, duration time.Duration) (string, error) {
	claims := Claims{
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
		Groups:   groups,
		Purpose:  purpose,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(secret string, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
