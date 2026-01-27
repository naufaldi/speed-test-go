package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/user/speed-test-go/pkg/types"
)

const serverListURL = "https://www.speedtest.net/api/js/servers?engine=js&limit=10"

// FetchServerList retrieves the list of available speed test servers
func FetchServerList(ctx context.Context) ([]*types.Server, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", serverListURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch server list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var servers []*types.Server
	if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
		return nil, fmt.Errorf("failed to decode server list: %w", err)
	}

	return servers, nil
}

// GetServerHost extracts the host from a server URL (for display purposes)
func GetServerHost(server *types.Server) string {
	url := server.URL

	// Remove protocol
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}

	// Remove trailing path
	if idx := strings.Index(url, "/"); idx > 0 {
		url = url[:idx]
	}

	return url
}

// GetServerBaseURL returns the base URL for HTTP requests (with protocol)
func GetServerBaseURL(server *types.Server) string {
	url := server.URL

	// Ensure protocol exists
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	// Remove trailing path
	if idx := strings.Index(url, "/speedtest/"); idx > 0 {
		url = url[:idx]
	}

	return url
}
