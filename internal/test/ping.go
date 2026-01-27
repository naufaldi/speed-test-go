package test

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/user/speed-test-go/internal/network"
	"github.com/user/speed-test-go/pkg/types"
)

// PingTest measures latency to a server
type PingTest struct {
	client      *http.Client
	numPings    int
	pingTimeout time.Duration
}

// NewPingTest creates a new ping test instance
func NewPingTest() *PingTest {
	return &PingTest{
		client:      network.NewHTTPClient(),
		numPings:    5,
		pingTimeout: 5 * time.Second,
	}
}

// Run executes the ping test and returns latency measurements
func (pt *PingTest) Run(ctx context.Context, serverURL string) ([]time.Duration, error) {
	latencyURL := fmt.Sprintf("%s/speedtest/latency.txt", serverURL)

	var latencies []time.Duration
	var mu sync.Mutex

	// Run pings concurrently with rate limiting
	sem := make(chan struct{}, 3) // Max 3 concurrent pings

	var wg sync.WaitGroup
	for i := 0; i < pt.numPings; i++ {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(pingNum int) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			select {
			case <-ctx.Done():
				return
			default:
			}

			start := time.Now()

			ctx, cancel := context.WithTimeout(ctx, pt.pingTimeout)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, "GET", latencyURL, nil)
			if err != nil {
				return
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

			resp, err := pt.client.Do(req)
			if err != nil {
				return
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				latency := time.Since(start)
				mu.Lock()
				latencies = append(latencies, latency)
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	return latencies, nil
}

// CalculateLatency calculates average latency and jitter from measurements
func CalculateLatency(latencies []time.Duration) (float64, float64) {
	if len(latencies) == 0 {
		return 0, 0
	}

	var sum float64
	for _, l := range latencies {
		sum += l.Seconds()
	}

	avg := sum / float64(len(latencies))

	// Calculate standard deviation (jitter)
	var varianceSum float64
	for _, l := range latencies {
		diff := l.Seconds() - avg
		varianceSum += diff * diff
	}

	stdDev := math.Sqrt(varianceSum / float64(len(latencies)))

	return avg * 1000, stdDev * 1000 // Convert to milliseconds
}

// RunPingTest is a convenience function to run a complete ping test
func RunPingTest(ctx context.Context, serverURL string) (*types.PingResult, error) {
	pt := NewPingTest()
	latencies, err := pt.Run(ctx, serverURL)
	if err != nil {
		return nil, err
	}

	// Return error if no successful pings
	if len(latencies) == 0 {
		return nil, fmt.Errorf("no successful pings to %s", serverURL)
	}

	latency, jitter := CalculateLatency(latencies)

	return &types.PingResult{
		Jitter:  jitter,
		Latency: latency,
	}, nil
}
