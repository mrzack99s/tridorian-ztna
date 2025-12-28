package utils

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"
)

var slugRegex = regexp.MustCompile("[^a-z0-9]+")

// GenerateSlug converts a string (like company name) into a URL-friendly slug.
func GenerateSlug(s string) string {
	s = strings.ToLower(s)
	s = slugRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")

	// Max length for a subdomain is 63, let's keep it safe
	if len(s) > 50 {
		s = s[:50]
	}
	return s
}

// GenerateRandomString produces a random string of fixed length using lowercase alphanumeric characters.
func GenerateRandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		if err != nil {
			// Fallback in unlikely case of error, though crypto/rand shouldn't typically fail
			return "random"
		}
		ret[i] = letterBytes[num.Int64()]
	}
	return string(ret)
}
