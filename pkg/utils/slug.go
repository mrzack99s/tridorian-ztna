package utils

import (
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
