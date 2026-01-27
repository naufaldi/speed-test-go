package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/user/speed-test-go/pkg/types"
)

const (
	defaultPingTimeout = 5 * time.Second
	maxConcurrentPings = 3
)

// ServerLatency holds the latency measurement for a server
type ServerLatency struct {
	Server   *types.Server
	Latency  time.Duration
	NumPings int
}

// FindServerByID searches for a server by its ID
func FindServerByID(servers []*types.Server, id string) *types.Server {
	for _, s := range servers {
		if s.ID == id {
			return s
		}
	}
	return nil
}

// SelectBestServerByPing selects the best server by pinging the top N closest servers
// and choosing the one with the lowest latency. If numServers is 0 or greater than
// available servers, it will ping all available servers.
func SelectBestServerByPing(ctx context.Context, servers []*types.Server, numServers int) (*types.Server, error) {
	if len(servers) == 0 {
		return nil, fmt.Errorf("no servers available")
	}

	// Limit numServers to available servers
	if numServers <= 0 || numServers > len(servers) {
		numServers = len(servers)
	}

	// Take the top N closest servers
	serversToTest := servers[:numServers]

	// Ping all servers concurrently
	latencies := pingServers(ctx, serversToTest)

	if len(latencies) == 0 {
		// If all pings failed, return the closest server
		// Don't log here as it will interfere with output
		return servers[0], fmt.Errorf("all ping attempts failed, using closest server")
	}

	// Find server with lowest average latency
	best := latencies[0]
	for _, sl := range latencies {
		if sl.Latency < best.Latency {
			best = sl
		}
	}

	return best.Server, nil
}

// pingServers pings multiple servers concurrently and returns their latencies
func pingServers(ctx context.Context, servers []*types.Server) []ServerLatency {
	client := &http.Client{
		Timeout: defaultPingTimeout,
	}

	var mu sync.Mutex
	var results []ServerLatency
	var wg sync.WaitGroup

	sem := make(chan struct{}, maxConcurrentPings)

	for _, srv := range servers {
		wg.Add(1)
		sem <- struct{}{}

		go func(s *types.Server) {
			defer wg.Done()
			defer func() { <-sem }()

			select {
			case <-ctx.Done():
				return
			default:
			}

			latency := measureServerLatency(ctx, client, s)
			if latency > 0 {
				mu.Lock()
				results = append(results, ServerLatency{
					Server:  s,
					Latency: latency,
				})
				mu.Unlock()
			}
		}(srv)
	}

	wg.Wait()
	return results
}

// measureServerLatency measures the latency to a single server
func measureServerLatency(ctx context.Context, client *http.Client, server *types.Server) time.Duration {
	latencyURL := fmt.Sprintf("%s/speedtest/latency.txt", server.URL)

	// Run 3 pings and take the average
	var latencies []time.Duration

	for i := 0; i < 3; i++ {
		select {
		case <-ctx.Done():
			return 0
		default:
		}

		start := time.Now()

		ctx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", latencyURL, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			latencies = append(latencies, time.Since(start))
		}
	}

	if len(latencies) == 0 {
		return 0
	}

	// Calculate average latency
	var sum time.Duration
	for _, l := range latencies {
		sum += l
	}

	return sum / time.Duration(len(latencies))
}

// GetLatencyResult returns a PingResult for a single server
func GetLatencyResult(ctx context.Context, server *types.Server) (*types.PingResult, error) {
	if server == nil {
		return nil, fmt.Errorf("server is nil")
	}

	client := &http.Client{
		Timeout: defaultPingTimeout,
	}

	latency := measureServerLatency(ctx, client, server)
	if latency == 0 {
		return nil, fmt.Errorf("failed to measure latency to server %s", server.ID)
	}

	return &types.PingResult{
		Latency: float64(latency.Milliseconds()),
		Jitter:  0, // Single measurement has no jitter
	}, nil
}
