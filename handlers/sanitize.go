package handlers

import (
	"regexp"
	"strings"
)

// SanitizeText removes basic HTML tags to prevent XSS
func SanitizeText(text string) string {
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	return strings.TrimSpace(text)
}

// IsValidImageURL checks if the URL is an HTTP/HTTPS link to prevent javascript: and data: injections
func IsValidImageURL(url string) bool {
	if url == "" {
		return true // Optional
	}
	url = strings.ToLower(strings.TrimSpace(url))
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}
	// Basic regex for valid domain/path
	matched, _ := regexp.MatchString(`^https?://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/\S*)?$`, url)
	return matched
}

// IsValidDiscordImageURL is a stricter check for Discord CDNs specifically used for proofs
func IsValidDiscordImageURL(url string) bool {
	if url == "" {
		return false
	}
	url = strings.ToLower(strings.TrimSpace(url))
	return strings.HasPrefix(url, "https://cdn.discordapp.com/") || strings.HasPrefix(url, "https://media.discordapp.net/")
}
