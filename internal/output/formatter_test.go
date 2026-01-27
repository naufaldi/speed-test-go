package output

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/user/speed-test-go/pkg/types"
)

func TestNewFormatter(t *testing.T) {
	f := NewFormatter(false, false, false)

	if f == nil {
		t.Fatal("Expected non-nil Formatter")
	}

	if f.useBytes {
		t.Error("Expected useBytes to be false")
	}

	if f.useJSON {
		t.Error("Expected useJSON to be false")
	}

	if f.useVerbose {
		t.Error("Expected useVerbose to be false")
	}
}

func TestNewFormatter_AllOptions(t *testing.T) {
	f := NewFormatter(true, true, true)

	if !f.useBytes {
		t.Error("Expected useBytes to be true")
	}

	if !f.useJSON {
		t.Error("Expected useJSON to be true")
	}

	if !f.useVerbose {
		t.Error("Expected useVerbose to be true")
	}
}

func TestFormatter_Format_HumanReadable(t *testing.T) {
	f := NewFormatter(false, false, false)

	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping: types.PingResult{
			Latency: 25.5,
			Jitter:  2.3,
		},
		Download: types.TransferResult{
			Bandwidth: 10000000, // 10 Mbps
			Bytes:     50000000,
			Elapsed:   5000,
		},
		Upload: types.TransferResult{
			Bandwidth: 5000000, // 5 Mbps
			Bytes:     25000000,
			Elapsed:   5000,
		},
	}

	output := f.Format(result)

	// Check that output contains expected values
	if !contains(output, "Ping") {
		t.Error("Expected output to contain 'Ping'")
	}

	if !contains(output, "Download") {
		t.Error("Expected output to contain 'Download'")
	}

	if !contains(output, "Upload") {
		t.Error("Expected output to contain 'Upload'")
	}

	if !contains(output, "25.5") {
		t.Error("Expected output to contain latency value 25.5")
	}

	// Should NOT contain JSON
	if contains(output, "{") {
		t.Error("Expected human-readable format, not JSON")
	}
}

func TestFormatter_Format_JSON(t *testing.T) {
	f := NewFormatter(false, true, false)

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
	}

	output := f.Format(result)

	// Should be valid JSON
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(output), &parsed)

	if err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}

	// Check that it contains expected keys
	if _, ok := parsed["ping"]; !ok {
		t.Error("Expected JSON to contain 'ping' key")
	}

	if _, ok := parsed["download"]; !ok {
		t.Error("Expected JSON to contain 'download' key")
	}

	if _, ok := parsed["upload"]; !ok {
		t.Error("Expected JSON to contain 'upload' key")
	}
}

func TestFormatter_Format_Verbose(t *testing.T) {
	f := NewFormatter(false, false, true)

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
			Sponsor:  "Test Sponsor",
			Distance: 150.5,
		},
	}

	output := f.Format(result)

	// Verbose mode should include server info
	if !contains(output, "Server") {
		t.Error("Expected verbose output to contain 'Server'")
	}

	if !contains(output, "test.example.com") {
		t.Error("Expected verbose output to contain server host")
	}

	if !contains(output, "Test Server") {
		t.Error("Expected verbose output to contain server name")
	}

	if !contains(output, "USA") {
		t.Error("Expected verbose output to contain country")
	}

	if !contains(output, "150.5") {
		t.Error("Expected verbose output to contain distance")
	}
}

func TestFormatter_Format_BytesMode(t *testing.T) {
	f := NewFormatter(true, false, false)

	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping: types.PingResult{
			Latency: 25.5,
			Jitter:  2.3,
		},
		Download: types.TransferResult{
			Bandwidth: 8388608, // 8 Mbps = 1 MB/s
			Bytes:     50000000,
			Elapsed:   5000,
		},
		Upload: types.TransferResult{
			Bandwidth: 4194304, // 4 Mbps = 0.5 MB/s
			Bytes:     25000000,
			Elapsed:   5000,
		},
	}

	output := f.Format(result)

	// Should show MBps instead of Mbps
	if !contains(output, "MBps") {
		t.Error("Expected bytes mode to contain 'MBps'")
	}

	if contains(output, "Mbps") {
		t.Error("Expected bytes mode to NOT contain 'Mbps'")
	}
}

func TestFormatter_Format_SpeedConversion(t *testing.T) {
	f := NewFormatter(false, false, false)

	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping:      types.PingResult{Latency: 25.5},
		Download:  types.TransferResult{Bandwidth: 1250000}, // 10 Mbps (1,250,000 bytes = 10 Mbps)
		Upload:    types.TransferResult{Bandwidth: 625000},  // 5 Mbps (625,000 bytes = 5 Mbps)
	}

	output := f.Format(result)

	// 10 Mbps = 10.00 Mbps
	if !contains(output, "10.00") {
		t.Error("Expected output to contain '10.00' for download speed")
	}

	// 5 Mbps = 5.00 Mbps
	if !contains(output, "5.00") {
		t.Error("Expected output to contain '5.00' for upload speed")
	}
}

func TestFormatter_FormatJSON_MarshalError(t *testing.T) {
	// This is a corner case - normally json.MarshalIndent won't fail
	// on our types, but we test the error handling path conceptually
	f := NewFormatter(false, true, false)

	// Even with unusual input, should not panic
	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping:      types.PingResult{Latency: 25.5},
		Download:  types.TransferResult{Bandwidth: 10000000},
		Upload:    types.TransferResult{Bandwidth: 5000000},
	}

	output := f.Format(result)

	// Should produce valid JSON
	if !contains(output, "{") || !contains(output, "}") {
		t.Error("Expected JSON output with braces")
	}
}

func TestFormatter_FormatError_JSON(t *testing.T) {
	f := NewFormatter(false, true, false)

	err := &testError{"test error message"}
	output := f.FormatError(err)

	// Should be valid JSON
	var parsed map[string]interface{}
	err2 := json.Unmarshal([]byte(output), &parsed)

	if err2 != nil {
		t.Errorf("Expected valid JSON error output, got: %s", output)
	}

	if _, ok := parsed["error"]; !ok {
		t.Error("Expected JSON error output to contain 'error' key")
	}
}

func TestFormatter_FormatError_Human(t *testing.T) {
	f := NewFormatter(false, false, false)

	err := &testError{"test error message"}
	output := f.FormatError(err)

	if !contains(output, "Error:") {
		t.Error("Expected human error format to contain 'Error:'")
	}

	if !contains(output, "test error message") {
		t.Error("Expected error message in output")
	}
}

func TestFormatter_Format_WithoutServer(t *testing.T) {
	f := NewFormatter(false, false, true)

	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping:      types.PingResult{Latency: 25.5},
		Download:  types.TransferResult{Bandwidth: 10000000},
		Upload:    types.TransferResult{Bandwidth: 5000000},
		Server:    nil, // No server info
	}

	output := f.Format(result)

	// Should still work without server info
	if !contains(output, "Ping") {
		t.Error("Expected output to contain 'Ping'")
	}
}

func TestFormatter_Format_ZeroValues(t *testing.T) {
	f := NewFormatter(false, false, false)

	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping:      types.PingResult{Latency: 0},
		Download:  types.TransferResult{Bandwidth: 0},
		Upload:    types.TransferResult{Bandwidth: 0},
	}

	output := f.Format(result)

	// Should handle zero values gracefully
	if !contains(output, "0.00") {
		t.Error("Expected output to contain zero values")
	}
}

func TestFormatter_Format_LargeValues(t *testing.T) {
	f := NewFormatter(false, false, false)

	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping:      types.PingResult{Latency: 999.9},
		Download:  types.TransferResult{Bandwidth: 125000000}, // 1 Gbps (125,000,000 bytes = 1000 Mbps)
		Upload:    types.TransferResult{Bandwidth: 62500000},  // 500 Mbps
	}

	output := f.Format(result)

	// Should format large values correctly (1000.00 Mbps)
	if !contains(output, "1000.00") {
		t.Error("Expected output to contain formatted large download speed")
	}
}

// Helper function to check substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// testError for testing error formatting
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestFormatSpeed_Mbps(t *testing.T) {
	// Test the formatSpeed helper function
	// formatSpeed converts bytes to Mbps by multiplying by 8 and dividing by 1,000,000
	testCases := []struct {
		bytesPerSecond int64
		expected       string
	}{
		{1250000, "10.00 Mbps"}, // 1,250,000 * 8 / 1,000,000 = 10 Mbps
		{125000, "1.00 Mbps"},   // 125,000 * 8 / 1,000,000 = 1 Mbps
		{12500, "0.10 Mbps"},    // 12,500 * 8 / 1,000,000 = 0.1 Mbps
		{1250000, "10.00 Mbps"},
	}

	for _, tc := range testCases {
		result := formatSpeed(tc.bytesPerSecond, false)
		if result != tc.expected {
			t.Errorf("formatSpeed(%d, false) = %s, expected %s", tc.bytesPerSecond, result, tc.expected)
		}
	}
}

func TestFormatSpeed_MBps(t *testing.T) {
	// Test the formatSpeed helper function with bytes mode
	testCases := []struct {
		bytesPerSecond int64
		expected       string
	}{
		{1048576, "1.00 MBps"},   // 1 MB/s
		{5242880, "5.00 MBps"},   // 5 MB/s
		{10485760, "10.00 MBps"}, // 10 MB/s
	}

	for _, tc := range testCases {
		result := formatSpeed(tc.bytesPerSecond, true)
		if result != tc.expected {
			t.Errorf("formatSpeed(%d, true) = %s, expected %s", tc.bytesPerSecond, result, tc.expected)
		}
	}
}

func TestFormatter_OutputFormat(t *testing.T) {
	f := NewFormatter(false, false, false)

	result := &types.SpeedTestResult{
		Timestamp: time.Now(),
		Ping:      types.PingResult{Latency: 25.5},
		Download:  types.TransferResult{Bandwidth: 10000000},
		Upload:    types.TransferResult{Bandwidth: 5000000},
	}

	output := f.Format(result)

	// Check that output has the expected format structure
	lines := 0
	for i := 0; i < len(output); i++ {
		if output[i] == '\n' {
			lines++
		}
	}

	if lines < 2 {
		t.Errorf("Expected at least 2 lines of output, got: %d", lines+1)
	}
}
