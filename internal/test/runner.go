package test

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/user/speed-test-go/internal/location"
	"github.com/user/speed-test-go/internal/server"
	"github.com/user/speed-test-go/internal/transfer"
	"github.com/user/speed-test-go/pkg/types"
)

// Runner orchestrates the complete speed test
type Runner struct {
	maxServers       int
	serverID         string
	numServersToTest int
}

// NewRunner creates a new test runner
func NewRunner() *Runner {
	return &Runner{
		maxServers:       5,
		numServersToTest: 5,
	}
}

// SetServerID sets the specific server ID to use
func (r *Runner) SetServerID(id string) {
	r.serverID = id
}

// SetNumServersToTest sets the number of closest servers to test for selection
func (r *Runner) SetNumServersToTest(n int) {
	if n > 0 {
		r.numServersToTest = n
	}
}

// Run executes the complete speed test
func (r *Runner) Run(ctx context.Context) (*types.SpeedTestResult, error) {
	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
	}

	// Step 1: Detect user location
	loc, err := location.DetectUserLocation(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect location: %w", err)
	}
	result.Interface = &types.InterfaceInfo{
		ExternalIP: loc.IP,
	}
	result.ISP = loc.ISP

	// Step 2: Fetch and sort servers
	servers, err := server.FetchServerList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch servers: %w", err)
	}

	userLat, _ := parseCoordinate(loc.Latitude)
	userLon, _ := parseCoordinate(loc.Longitude)
	server.CalculateServerDistances(servers, userLat, userLon)

	// Sort by distance
	server.SortServersByDistance(servers)

	// Step 3: Select best server
	var bestServer *types.Server
	var selErr error

	if r.serverID != "" {
		// Use specified server ID
		bestServer = server.FindServerByID(servers, r.serverID)
		if bestServer == nil {
			return nil, fmt.Errorf("server with ID %s not found", r.serverID)
		}
	} else {
		// Auto-select best server by pinging closest servers
		bestServer, selErr = server.SelectBestServerByPing(ctx, servers, r.numServersToTest)
		if selErr != nil {
			// Fall back to closest server
			bestServer = servers[0]
		}
	}

	serverURL := server.GetServerBaseURL(bestServer)

	// Step 4: Run ping test
	pingResult, err := RunPingTest(ctx, serverURL)
	if err != nil {
		return nil, fmt.Errorf("ping test failed: %w", err)
	}
	result.Ping = *pingResult

	// Step 5: Run download test
	downloadResult, err := transfer.RunSimpleDownloadTest(ctx, serverURL)
	if err != nil {
		return nil, fmt.Errorf("download test failed: %w", err)
	}
	result.Download = types.TransferResult{
		Bandwidth: downloadResult.Bandwidth,
		Bytes:     downloadResult.Bytes,
		Elapsed:   downloadResult.Elapsed.Milliseconds(),
	}

	// Step 6: Run upload test
	uploadResult, err := transfer.RunSimpleUploadTest(ctx, serverURL)
	if err != nil {
		return nil, fmt.Errorf("upload test failed: %w", err)
	}
	result.Upload = types.TransferResult{
		Bandwidth: uploadResult.Bandwidth,
		Bytes:     uploadResult.Bytes,
		Elapsed:   uploadResult.Elapsed.Milliseconds(),
	}

	// Step 7: Populate server info
	result.Server = &types.ServerInfo{
		ID:       bestServer.ID,
		Host:     bestServer.Host,
		Name:     bestServer.Name,
		Country:  bestServer.Country,
		Sponsor:  bestServer.Sponsor,
		Distance: bestServer.Distance,
	}

	return result, nil
}

func parseCoordinate(f float64) (float64, float64) {
	return f, f
}

func parseCoordinateString(coord string) (float64, error) {
	return strconv.ParseFloat(coord, 64)
}
