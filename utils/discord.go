package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// IsDiscordURLExpired checks if a Discord CDN URL is expired or expires within the buffer duration.
// Returns false if it's not a Discord URL or if it can't be parsed.
func IsDiscordURLExpired(rawURL string) bool {
	if !strings.HasPrefix(rawURL, "https://cdn.discordapp.com/") && !strings.HasPrefix(rawURL, "https://media.discordapp.net/") {
		return false // Not a discord URL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	ex := u.Query().Get("ex")
	if ex == "" {
		return false // No expiry param, maybe old or public?
	}

	timestamp, err := strconv.ParseInt(ex, 16, 64)
	if err != nil {
		return false
	}

	expiryTime := time.Unix(timestamp, 0)
	// If it expires in less than 1 hour, consider it expired to be safe
	return time.Now().Add(1 * time.Hour).After(expiryTime)
}

type RefreshRequest struct {
	AttachmentURLs []string `json:"attachment_urls"`
}

type RefreshResponse struct {
	RefreshedURLs []struct {
		Original  string `json:"original"`
		Refreshed string `json:"refreshed"`
	} `json:"refreshed_urls"`
}

// RefreshDiscordURLs takes a slice of expired Discord URLs and refreshes them via Discord API.
// Requires DISCORD_BOT_TOKEN environment variable.
func RefreshDiscordURLs(urls []string) (map[string]string, error) {
	if len(urls) == 0 {
		return nil, nil
	}

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_BOT_TOKEN is not set")
	}

	reqBody, _ := json.Marshal(RefreshRequest{AttachmentURLs: urls})

	req, err := http.NewRequest("POST", "https://discord.com/api/v10/attachments/refresh-urls", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discord API error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var refreshResp RefreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&refreshResp); err != nil {
		return nil, err
	}

	resultMap := make(map[string]string)
	for _, item := range refreshResp.RefreshedURLs {
		resultMap[item.Original] = item.Refreshed
	}

	return resultMap, nil
}
