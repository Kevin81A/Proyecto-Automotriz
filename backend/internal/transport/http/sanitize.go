package http

import (
	"regexp"
	"strings"
)

var (
	scriptTagRegex = regexp.MustCompile(`(?i)<script\b[^>]*>[\s\S]*?</script>`)
	htmlTagRegex   = regexp.MustCompile(`(?i)<[^>]+>`)
)

// sanitizeString removes HTML script blocks and any HTML tags, and trims whitespace.
func sanitizeString(val string) string {
	cleaned := scriptTagRegex.ReplaceAllString(val, "")
	cleaned = htmlTagRegex.ReplaceAllString(cleaned, "")
	return strings.TrimSpace(cleaned)
}
