package location

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/user/speed-test-go/pkg/types"
)

const configURL = "http://speedtest.net/speedtest-config.php"

// DetectUserLocation detects the user's location based on their IP
func DetectUserLocation(ctx context.Context) (*types.UserLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", configURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch location: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var settings types.Settings
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&settings); err != nil {
		return nil, fmt.Errorf("failed to decode location: %w", err)
	}

	return &types.UserLocation{
		IP:        settings.Client.IP,
		Latitude:  settings.Client.Latitude,
		Longitude: settings.Client.Longitude,
		ISP:       settings.Client.ISP,
	}, nil
}
