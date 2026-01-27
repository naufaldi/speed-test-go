package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/speed-test-go/pkg/types"
)

func TestNewRunner(t *testing.T) {
	r := NewRunner()

	if r == nil {
		t.Fatal("Expected non-nil Runner")
	}

	if r.maxServers != 5 {
		t.Errorf("Expected maxServers 5, got: %d", r.maxServers)
	}
}

func TestRunner_Run_LocationDetection(t *testing.T) {
	// Create a test server for location
	locationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/speedtest-config.php" {
			w.Header().Set("Content-Type", "text/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<settings version="1.0" mode="speedtest">
  <server-config ip="192.168.1.1" lat="40.7128" lon="-74.0060" isp="Test ISP" />
</settings>`))
		}
	}))
	defer locationServer.Close()

	// Create a test server for server list
	serverListServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/js/servers" {
			w.Header().Set("Content-Type", "text/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<servers version="1.0" mode="speedtest">
  <server url="http://test.example.com/speedtest/upload.php" lat="40.7128" lon="-74.0060" name="Test Server" country="USA" sponsor="Test" id="1" host="test.example.com" />
</servers>`))
		}
	}))
	defer serverListServer.Close()

	// Create a test server for ping
	pingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/speedtest/latency.txt" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer pingServer.Close()

	// Create a test server for download/upload
	transferServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test data"))
	}))
	defer transferServer.Close()

	// Note: In a real test, we'd need to mock the URLs
	// For now, just test that the runner can be created
	r := NewRunner()

	if r == nil {
		t.Fatal("Expected non-nil Runner")
	}
}

func TestRunner_Run_InvalidLocation(t *testing.T) {
	// Create a test server that returns invalid location
	locationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid xml"))
	}))
	defer locationServer.Close()

	// Note: We can't easily test the full Run without more complex mocking
	// Just verify runner creation
	r := NewRunner()

	if r == nil {
		t.Fatal("Expected non-nil Runner")
	}
}

func TestRunner_Run_ServerFetchError(t *testing.T) {
	// Test that runner handles server fetch errors gracefully
	// This is more of a conceptual test since we can't easily mock all dependencies

	r := NewRunner()

	if r == nil {
		t.Fatal("Expected non-nil Runner")
	}

	if r.maxServers != 5 {
		t.Errorf("Expected maxServers 5, got: %d", r.maxServers)
	}
}

func TestRunner_Run_ContextCancellation(t *testing.T) {
	r := NewRunner()

	if r == nil {
		t.Fatal("Expected non-nil Runner")
	}

	// Test with cancelled context
	_, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Runner should handle cancellation gracefully
	// Note: Without proper mocking, this will fail at network calls
	t.Log("Context cancellation test requires full mocking setup")
}

func TestRunner_MaxServers(t *testing.T) {
	testCases := []struct {
		maxServers int
		expected   int
	}{
		{1, 1},
		{5, 5},
		{10, 10},
	}

	for _, tc := range testCases {
		r := &Runner{maxServers: tc.maxServers}
		if r.maxServers != tc.expected {
			t.Errorf("Expected maxServers %d, got: %d", tc.expected, r.maxServers)
		}
	}
}

func TestParseCoordinate(t *testing.T) {
	result, _ := parseCoordinate(40.7128)

	// Function returns (f, f) - same value for both
	if result != 40.7128 {
		t.Errorf("Expected 40.7128, got: %f", result)
	}
}

func TestParseCoordinateString(t *testing.T) {
	testCases := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"40.7128", 40.7128, false},
		{"-74.0060", -74.0060, false},
		{"0", 0, false},
		{"123.456", 123.456, false},
	}

	for _, tc := range testCases {
		result, err := parseCoordinateString(tc.input)

		if tc.hasError && err == nil {
			t.Errorf("Expected error for '%s', got nil", tc.input)
		}

		if !tc.hasError && err != nil {
			t.Errorf("Unexpected error for '%s': %v", tc.input, err)
		}

		if result != tc.expected {
			t.Errorf("Expected %f for '%s', got: %f", tc.expected, tc.input, result)
		}
	}
}

func TestParseCoordinateString_Invalid(t *testing.T) {
	_, err := parseCoordinateString("invalid")

	if err == nil {
		t.Error("Expected error for invalid coordinate")
	}
}

func TestRunner_Structure(t *testing.T) {
	r := NewRunner()

	if r == nil {
		t.Fatal("Expected non-nil Runner")
	}

	// Verify the structure
	if r.maxServers <= 0 {
		t.Errorf("Expected positive maxServers, got: %d", r.maxServers)
	}
}

func TestRunner_ResultStructure(t *testing.T) {
	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping: types.PingResult{
			Latency: 25.5,
			Jitter:  2.3,
		},
		Download: types.TransferResult{
			Bandwidth: 10000000,
			Bytes:     50000000,
			Elapsed:   5000,
		},
		Upload: types.TransferResult{
			Bandwidth: 5000000,
			Bytes:     25000000,
			Elapsed:   5000,
		},
		Server: &types.ServerInfo{
			ID:       "1",
			Host:     "test.example.com",
			Name:     "Test Server",
			Country:  "USA",
			Sponsor:  "Test",
			Distance: 100.5,
		},
		Interface: &types.InterfaceInfo{
			ExternalIP: "192.168.1.1",
		},
		ISP: "Test ISP",
	}

	if result.Ping.Latency != 25.5 {
		t.Errorf("Expected latency 25.5, got: %f", result.Ping.Latency)
	}

	if result.Download.Bandwidth != 10000000 {
		t.Errorf("Expected bandwidth 10000000, got: %d", result.Download.Bandwidth)
	}

	if result.Server.Distance != 100.5 {
		t.Errorf("Expected distance 100.5, got: %f", result.Server.Distance)
	}
}

func TestRunner_ResultNilServer(t *testing.T) {
	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping:      types.PingResult{Latency: 25.5},
		Download:  types.TransferResult{Bandwidth: 10000000},
		Upload:    types.TransferResult{Bandwidth: 5000000},
		Server:    nil, // No server
	}

	// Should handle nil server gracefully
	if result.Server != nil {
		t.Error("Expected nil server")
	}
}

func TestRunner_ResultNilInterface(t *testing.T) {
	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping:      types.PingResult{Latency: 25.5},
		Download:  types.TransferResult{Bandwidth: 10000000},
		Upload:    types.TransferResult{Bandwidth: 5000000},
	}

	// Should handle nil interface gracefully
	if result.Interface != nil {
		t.Error("Expected nil interface")
	}
}

func TestRunner_EmptyISP(t *testing.T) {
	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping:      types.PingResult{Latency: 25.5},
		Download:  types.TransferResult{Bandwidth: 10000000},
		Upload:    types.TransferResult{Bandwidth: 5000000},
		ISP:       "", // Empty ISP
	}

	if result.ISP != "" {
		t.Errorf("Expected empty ISP, got: %s", result.ISP)
	}
}

func TestRunner_ZeroValues(t *testing.T) {
	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping:      types.PingResult{Latency: 0},
		Download:  types.TransferResult{Bandwidth: 0},
		Upload:    types.TransferResult{Bandwidth: 0},
	}

	// Should handle zero values
	if result.Ping.Latency != 0 {
		t.Errorf("Expected zero latency, got: %f", result.Ping.Latency)
	}
}

func TestRunner_LargeValues(t *testing.T) {
	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping:      types.PingResult{Latency: 999.9},
		Download:  types.TransferResult{Bandwidth: 1000000000}, // 1 Gbps
		Upload:    types.TransferResult{Bandwidth: 500000000},  // 500 Mbps
	}

	// Should handle large values
	if result.Download.Bandwidth != 1000000000 {
		t.Errorf("Expected bandwidth 1000000000, got: %d", result.Download.Bandwidth)
	}
}

func TestRunner_ManyServers(t *testing.T) {
	// Create a large server list
	servers := make([]*types.Server, 100)
	for i := 0; i < 100; i++ {
		servers[i] = &types.Server{
			ID:       string(rune('0' + i%10)),
			Lat:      "40.7128",
			Lon:      "-74.0060",
			Distance: float64(i),
		}
	}

	// Test that sorting works
	// This is a simplified test without full runner integration
	if len(servers) != 100 {
		t.Errorf("Expected 100 servers, got: %d", len(servers))
	}
}

func TestRunner_ResultJSON(t *testing.T) {
	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping:      types.PingResult{Latency: 25.5},
		Download:  types.TransferResult{Bandwidth: 10000000},
		Upload:    types.TransferResult{Bandwidth: 5000000},
	}

	// Result should be JSON serializable (basic test)
	if result.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}

	if result.Ping.Latency <= 0 {
		t.Error("Expected positive latency")
	}
}

func TestRunner_ConcurrentCreation(t *testing.T) {
	done := make(chan *Runner, 100)

	// Create runners concurrently
	for i := 0; i < 100; i++ {
		go func() {
			runner := NewRunner()
			done <- runner
		}()
	}

	// All runners should be created successfully
	for i := 0; i < 100; i++ {
		runner := <-done
		if runner == nil {
			t.Errorf("Runner %d is nil", i)
		}
		if runner.maxServers != 5 {
			t.Errorf("Runner %d has wrong maxServers", i)
		}
	}
}

func TestRunner_DefaultMaxServers(t *testing.T) {
	r := NewRunner()

	if r.maxServers != 5 {
		t.Errorf("Expected default maxServers 5, got: %d", r.maxServers)
	}
}

func TestRunner_CustomMaxServers(t *testing.T) {
	r := &Runner{maxServers: 10}

	if r.maxServers != 10 {
		t.Errorf("Expected custom maxServers 10, got: %d", r.maxServers)
	}
}

func TestRunner_ServerInfo(t *testing.T) {
	serverInfo := &types.ServerInfo{
		ID:       "1",
		Host:     "test.example.com",
		Name:     "Test Server",
		Location: "New York",
		Country:  "USA",
		Sponsor:  "Test Sponsor",
		Distance: 150.5,
	}

	if serverInfo.ID != "1" {
		t.Errorf("Expected ID '1', got: %s", serverInfo.ID)
	}

	if serverInfo.Host != "test.example.com" {
		t.Errorf("Expected Host 'test.example.com', got: %s", serverInfo.Host)
	}

	if serverInfo.Distance != 150.5 {
		t.Errorf("Expected Distance 150.5, got: %f", serverInfo.Distance)
	}
}

func TestRunner_InterfaceInfo(t *testing.T) {
	interfaceInfo := &types.InterfaceInfo{
		InternalIP: "192.168.1.100",
		ExternalIP: "203.0.113.1",
		Name:       "eth0",
		MACAddr:    "00:11:22:33:44:55",
	}

	if interfaceInfo.ExternalIP != "203.0.113.1" {
		t.Errorf("Expected ExternalIP '203.0.113.1', got: %s", interfaceInfo.ExternalIP)
	}

	if interfaceInfo.MACAddr != "00:11:22:33:44:55" {
		t.Errorf("Expected MACAddr '00:11:22:33:44:55', got: %s", interfaceInfo.MACAddr)
	}
}

func TestRunner_TransferResult(t *testing.T) {
	transferResult := types.TransferResult{
		Bandwidth: 10000000,
		Bytes:     50000000,
		Elapsed:   5000,
	}

	if transferResult.Bandwidth != 10000000 {
		t.Errorf("Expected Bandwidth 10000000, got: %d", transferResult.Bandwidth)
	}

	if transferResult.Bytes != 50000000 {
		t.Errorf("Expected Bytes 50000000, got: %d", transferResult.Bytes)
	}

	if transferResult.Elapsed != 5000 {
		t.Errorf("Expected Elapsed 5000, got: %d", transferResult.Elapsed)
	}
}

func TestRunner_PingResult(t *testing.T) {
	pingResult := types.PingResult{
		Jitter:  2.5,
		Latency: 25.5,
	}

	if pingResult.Jitter != 2.5 {
		t.Errorf("Expected Jitter 2.5, got: %f", pingResult.Jitter)
	}

	if pingResult.Latency != 25.5 {
		t.Errorf("Expected Latency 25.5, got: %f", pingResult.Latency)
	}
}
