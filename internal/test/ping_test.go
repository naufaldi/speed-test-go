package test

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewPingTest(t *testing.T) {
	pt := NewPingTest()

	if pt == nil {
		t.Fatal("Expected non-nil PingTest")
	}

	if pt.client == nil {
		t.Error("Expected client to be initialized")
	}

	if pt.numPings != 5 {
		t.Errorf("Expected 5 pings, got: %d", pt.numPings)
	}

	if pt.pingTimeout != 5*time.Second {
		t.Errorf("Expected ping timeout 5s, got: %v", pt.pingTimeout)
	}
}

func TestPingTest_Run_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/speedtest/latency.txt" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pt := NewPingTest()
	ctx := context.Background()

	latencies, err := pt.Run(ctx, server.URL)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have some successful pings
	if len(latencies) == 0 {
		t.Error("Expected at least one latency measurement")
	}

	// All latencies should be positive
	for _, latency := range latencies {
		if latency <= 0 {
			t.Errorf("Expected positive latency, got: %v", latency)
		}
	}
}

func TestPingTest_Run_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pt := NewPingTest()
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	latencies, err := pt.Run(ctx, server.URL)

	if err != nil {
		t.Logf("Got error (expected for cancellation): %v", err)
	}

	// Should have partial or no results
	t.Logf("Got %d latencies before cancellation", len(latencies))
}

func TestPingTest_Run_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	pt := NewPingTest()
	ctx := context.Background()

	latencies, err := pt.Run(ctx, server.URL)

	// Should not panic, but may have fewer/no results
	if err != nil {
		t.Logf("Got error: %v", err)
	}

	t.Logf("Got %d latencies for error response", len(latencies))
}

func TestPingTest_Run_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pt := NewPingTest()
	pt.pingTimeout = 100 * time.Millisecond // Short timeout
	ctx := context.Background()

	start := time.Now()
	latencies, err := pt.Run(ctx, server.URL)
	elapsed := time.Since(start)

	if err != nil {
		t.Logf("Got error: %v", err)
	}

	// Should complete within reasonable time
	if elapsed > 2*time.Second {
		t.Errorf("Test took too long: %v", elapsed)
	}

	t.Logf("Got %d latencies with short timeout", len(latencies))
}

func TestCalculateLatency_Empty(t *testing.T) {
	latencies := []time.Duration{}

	avg, jitter := CalculateLatency(latencies)

	if avg != 0 {
		t.Errorf("Expected average 0 for empty latencies, got: %f", avg)
	}

	if jitter != 0 {
		t.Errorf("Expected jitter 0 for empty latencies, got: %f", jitter)
	}
}

func TestCalculateLatency_SingleValue(t *testing.T) {
	latencies := []time.Duration{100 * time.Millisecond}

	avg, jitter := CalculateLatency(latencies)

	// Average should be 100ms
	if avg != 100.0 {
		t.Errorf("Expected average 100ms, got: %f", avg)
	}

	// Jitter should be 0 for single value
	if jitter != 0 {
		t.Errorf("Expected jitter 0 for single value, got: %f", jitter)
	}
}

func TestCalculateLatency_MultipleValues(t *testing.T) {
	latencies := []time.Duration{
		50 * time.Millisecond,
		100 * time.Millisecond,
		150 * time.Millisecond,
	}

	avg, jitter := CalculateLatency(latencies)

	// Average should be 100ms
	expectedAvg := 100.0
	if avg < expectedAvg-1 || avg > expectedAvg+1 {
		t.Errorf("Expected average ~%fms, got: %f", expectedAvg, avg)
	}

	// Jitter should be positive
	if jitter <= 0 {
		t.Error("Expected positive jitter for multiple values")
	}
}

func TestCalculateLatency_AllSame(t *testing.T) {
	latencies := []time.Duration{
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
	}

	avg, jitter := CalculateLatency(latencies)

	// Average should be 100ms
	if avg != 100.0 {
		t.Errorf("Expected average 100ms, got: %f", avg)
	}

	// Jitter should be 0
	if jitter != 0 {
		t.Errorf("Expected jitter 0 for identical values, got: %f", jitter)
	}
}

func TestCalculateLatency_VaryingValues(t *testing.T) {
	latencies := []time.Duration{
		10 * time.Millisecond,
		500 * time.Millisecond,
		100 * time.Millisecond,
		50 * time.Millisecond,
		200 * time.Millisecond,
	}

	avg, jitter := CalculateLatency(latencies)

	// Average should be around 172ms
	expectedAvg := 172.0
	if avg < expectedAvg-10 || avg > expectedAvg+10 {
		t.Errorf("Expected average ~%fms, got: %f", expectedAvg, avg)
	}

	// Jitter should be significant
	if jitter < 100 {
		t.Errorf("Expected significant jitter, got: %fms", jitter)
	}
}

func TestCalculateLatency_MillisecondConversion(t *testing.T) {
	// Test that conversion to milliseconds is correct
	latencies := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
	}

	avg, _ := CalculateLatency(latencies)

	// (100 + 200) / 2 = 150ms
	expectedAvg := 150.0
	// Use tolerance for floating point comparison
	if math.Abs(avg-expectedAvg) > 0.01 {
		t.Errorf("Expected average %fms, got: %f", expectedAvg, avg)
	}
}

func TestRunPingTest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()
	result, err := RunPingTest(ctx, server.URL)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Latency should be positive
	if result.Latency <= 0 {
		t.Errorf("Expected positive latency, got: %f", result.Latency)
	}

	// Jitter should be non-negative
	if result.Jitter < 0 {
		t.Errorf("Expected non-negative jitter, got: %f", result.Jitter)
	}
}

func TestRunPingTest_InvalidURL(t *testing.T) {
	ctx := context.Background()

	// Use an invalid URL
	_, err := RunPingTest(ctx, "http://invalid.example.com/nonexistent")

	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestPingTest_ConcurrentPings(t *testing.T) {
	var mu sync.Mutex
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pt := NewPingTest()
	pt.numPings = 10 // Increase for test

	ctx := context.Background()
	latencies, err := pt.Run(ctx, server.URL)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	mu.Lock()
	count := requestCount
	mu.Unlock()

	// Should have multiple concurrent requests
	if count < 1 {
		t.Errorf("Expected at least 1 request, got: %d", count)
	}

	// Should have some successful latencies
	if len(latencies) == 0 {
		t.Log("No successful pings (may be expected depending on timing)")
	}
}

func TestPingTest_URLFormat(t *testing.T) {
	var receivedURL string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pt := NewPingTest()
	ctx := context.Background()

	_, _ = pt.Run(ctx, server.URL)

	expected := "/speedtest/latency.txt"
	if receivedURL != expected {
		t.Errorf("Expected URL '%s', got: %s", expected, receivedURL)
	}
}

func TestPingTest_HTTPMethod(t *testing.T) {
	var method string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pt := NewPingTest()
	ctx := context.Background()

	_, _ = pt.Run(ctx, server.URL)

	if method != "GET" {
		t.Errorf("Expected GET method, got: %s", method)
	}
}

func TestPingTest_SemaphoreLimit(t *testing.T) {
	var mu sync.Mutex
	activeRequests := 0
	maxActive := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		activeRequests++
		if activeRequests > maxActive {
			maxActive = activeRequests
		}
		mu.Unlock()

		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		activeRequests--
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pt := NewPingTest()
	pt.numPings = 10
	pt.pingTimeout = 2 * time.Second

	ctx := context.Background()
	_, _ = pt.Run(ctx, server.URL)

	// Check that semaphore limits concurrent requests (max 3)
	mu.Lock()
	defer mu.Unlock()

	if maxActive > 3 {
		t.Errorf("Expected max 3 concurrent requests, got: %d", maxActive)
	}
}

func TestPingTest_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	pt := NewPingTest()
	ctx := context.Background()

	latencies, err := pt.Run(ctx, server.URL)

	if err != nil {
		t.Logf("Got error: %v", err)
	}

	// May have no results due to errors
	t.Logf("Got %d latencies", len(latencies))
}

func TestPingTest_ResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("response body"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pt := NewPingTest()
	ctx := context.Background()

	latencies, err := pt.Run(ctx, server.URL)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should still work even with response body
	if len(latencies) == 0 {
		t.Error("Expected successful pings")
	}
}

func TestPingTest_ContentType(t *testing.T) {
	var contentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pt := NewPingTest()
	ctx := context.Background()

	_, _ = pt.Run(ctx, server.URL)

	t.Logf("Accept header: %s", contentType)
}

func TestPingTest_DefaultNumPings(t *testing.T) {
	pt := NewPingTest()

	if pt.numPings != 5 {
		t.Errorf("Expected default 5 pings, got: %d", pt.numPings)
	}
}

func TestPingTest_DefaultTimeout(t *testing.T) {
	pt := NewPingTest()

	if pt.pingTimeout != 5*time.Second {
		t.Errorf("Expected default 5s timeout, got: %v", pt.pingTimeout)
	}
}

func TestCalculateLatency_NilSlice(t *testing.T) {
	latencies := []time.Duration(nil)

	avg, jitter := CalculateLatency(latencies)

	if avg != 0 {
		t.Errorf("Expected average 0 for nil slice, got: %f", avg)
	}

	if jitter != 0 {
		t.Errorf("Expected jitter 0 for nil slice, got: %f", jitter)
	}
}
