package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/user/speed-test-go/pkg/types"
)

// Formatter handles output formatting
type Formatter struct {
	useBytes   bool
	useJSON    bool
	useVerbose bool
}

// NewFormatter creates a new formatter
func NewFormatter(useBytes, useJSON, useVerbose bool) *Formatter {
	return &Formatter{
		useBytes:   useBytes,
		useJSON:    useJSON,
		useVerbose: useVerbose,
	}
}

// Format formats the test result for output
func (f *Formatter) Format(result *types.SpeedTestResult) string {
	if f.useJSON {
		return f.formatJSON(result)
	}
	return f.formatHuman(result)
}

func (f *Formatter) formatJSON(result *types.SpeedTestResult) string {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to format result: %v"}`, err)
	}
	return string(data)
}

func (f *Formatter) formatHuman(result *types.SpeedTestResult) string {
	var sb strings.Builder

	// Format speed values
	downloadStr := formatSpeed(result.Download.Bandwidth, f.useBytes)
	uploadStr := formatSpeed(result.Upload.Bandwidth, f.useBytes)
	pingStr := fmt.Sprintf("%.1f ms", result.Ping.Latency)

	// Output format matching sindresorhus/speed-test
	sb.WriteString(fmt.Sprintf("      Ping %s\n", pingStr))
	sb.WriteString(fmt.Sprintf("  Download %s\n", downloadStr))
	sb.WriteString(fmt.Sprintf("    Upload %s\n", uploadStr))

	// Verbose mode - server information
	if f.useVerbose && result.Server != nil {
		sb.WriteString(fmt.Sprintf("\n"))
		sb.WriteString(fmt.Sprintf("    Server   %s\n", result.Server.Host))
		sb.WriteString(fmt.Sprintf("  Location   %s (%s)\n", result.Server.Name, result.Server.Country))
		sb.WriteString(fmt.Sprintf("  Distance   %.1f km\n", result.Server.Distance))
	}

	return sb.String()
}

// formatSpeed formats a speed value in Mbps or MB/s
func formatSpeed(bytesPerSecond int64, useBytes bool) string {
	if useBytes {
		// Convert to MB/s
		mbPerSecond := float64(bytesPerSecond) / (1024 * 1024)
		return fmt.Sprintf("%.2f MBps", mbPerSecond)
	}
	// Convert to Mbps
	mbps := float64(bytesPerSecond) * 8 / (1000 * 1000)
	return fmt.Sprintf("%.2f Mbps", mbps)
}

// FormatError formats an error message
func (f *Formatter) FormatError(err error) string {
	if f.useJSON {
		return fmt.Sprintf(`{"error": %q}`, err.Error())
	}
	return fmt.Sprintf("Error: %s\n", err.Error())
}
