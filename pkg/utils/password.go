package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
)

func GenerateSecureToken() string {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("trivpn_agent_%s", base64.RawURLEncoding.EncodeToString(b))
}

func GenerateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "password123" // Fallback
		}
		ret[i] = charset[num.Int64()]
	}
	return string(ret)
}
