package utils

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
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
	Email    string       `json:"email"`
	TenantID string       `json:"tenant_id"`
	Role     string       `json:"role,omitempty"`
	Groups   []string     `json:"groups,omitempty"` // Google Workspace groups
	OS       string       `json:"os,omitempty"`
	Purpose  TokenPurpose `json:"purpose"`
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token signed with EdDSA
// privateKey should be an ed25519.PrivateKey
func GenerateToken(privateKey interface{}, purpose TokenPurpose, userID, email, tenantID, role string, groups []string, os string, duration time.Duration) (string, error) {
	claims := Claims{
		UserID:   userID,
		Email:    email,
		TenantID: tenantID,
		Role:     role,
		Groups:   groups,
		OS:       os,
		Purpose:  purpose,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(privateKey)
}

// ParseToken parses and verifies a JWT token using EdDSA
// publicKey should be an ed25519.PublicKey
func ParseToken(publicKey interface{}, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// Key Management Helpers

func GenerateEd25519Key() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

func PrivateKeyToPEM(privateKey ed25519.PrivateKey) string {
	b, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return ""
	}
	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	}
	return string(pem.EncodeToMemory(block))
}

func PublicKeyToPEM(publicKey ed25519.PublicKey) string {
	b, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return ""
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}
	return string(pem.EncodeToMemory(block))
}

func ParseEdPrivateKeyFromPEM(pemStr string) (interface{}, error) {
	return jwt.ParseEdPrivateKeyFromPEM([]byte(pemStr))
}

func ParseEdPublicKeyFromPEM(pemStr string) (interface{}, error) {
	return jwt.ParseEdPublicKeyFromPEM([]byte(pemStr))
}
