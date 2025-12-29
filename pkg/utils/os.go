package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// secretMapping maps environment variable names to their CSI mount paths
var secretMapping = map[string]string{
	"ZTNA_PRIVATE_KEY": "/mnt/secrets-store/private-key",
	"ZTNA_PUBLIC_KEY":  "/mnt/secrets-store/public-key",
	"DB_USER":          "/mnt/db-secrets/username",
	"DB_PASSWORD":      "/mnt/db-secrets/password",
	"DB_NAME":          "/mnt/db-secrets/database",
	"DB_HOST":          "/mnt/db-secrets/host",
	"CACHE_PASSWORD":   "/mnt/cache-secrets/password",
}

func GetEnv(key, fallback string) string {
	// 1. Try to read from CSI mount path if it's a known secret
	if mountPath, ok := secretMapping[key]; ok {
		if content, err := os.ReadFile(mountPath); err == nil {
			return strings.TrimSpace(string(content))
		}
	}

	// 2. Try to read from Docker secret path (standard /run/secrets) as a secondary fallback
	dockerSecretPath := filepath.Join("/run/secrets", strings.ToLower(key))
	if content, err := os.ReadFile(dockerSecretPath); err == nil {
		return strings.TrimSpace(string(content))
	}

	// 3. Fallback to Environment Variable
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
