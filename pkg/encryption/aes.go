package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/chacha20poly1305"
)

func Encrypt(key, plaintext, aad []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, aad)
	out := append(nonce, ciphertext...)
	return out, nil
}

func Decrypt(key, frame, aad []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	nonceSize := aead.NonceSize()
	if len(frame) < nonceSize {
		return nil, errors.New("invalid frame")
	}

	nonce := frame[:nonceSize]
	ciphertext := frame[nonceSize:]

	plaintext, err := aead.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns a base64 encoded string
func EncryptString(plaintext string, key string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	// Key must be 32 bytes for chacha20poly1305
	fixedKey := adjustKey(key)
	encrypted, err := Encrypt(fixedKey, []byte(plaintext), nil)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptString decrypts a base64 encoded string and returns the plaintext
func DecryptString(cryptoText string, key string) (string, error) {
	if cryptoText == "" {
		return "", nil
	}
	fixedKey := adjustKey(key)
	decoded, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}
	decrypted, err := Decrypt(fixedKey, decoded, nil)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

func adjustKey(key string) []byte {
	k := make([]byte, 32)
	copy(k, key)
	return k
}
