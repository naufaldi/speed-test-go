package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/user/speed-test-go/internal/transfer"
	"github.com/user/speed-test-go/pkg/types"
)

// ProgressReporter handles real-time progress display
type ProgressReporter struct {
	formatter  *Formatter
	lastOutput string
	updateFreq time.Duration
	state      types.OutputState
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(formatter *Formatter) *ProgressReporter {
	return &ProgressReporter{
		formatter:  formatter,
		updateFreq: 50 * time.Millisecond,
		state:      types.StateIdle,
	}
}

// SetState updates the current test state
func (pr *ProgressReporter) SetState(state types.OutputState) {
	pr.state = state
}

// ReportPingProgress reports ping test progress
func (pr *ProgressReporter) ReportPingProgress(latency float64) {
	output := pr.formatProgressOutput(func() string {
		return fmt.Sprintf("      Ping %.1f ms", latency)
	})
	pr.printOutput(output)
}

// ReportDownloadProgress reports download test progress
func (pr *ProgressReporter) ReportDownloadProgress(progress transfer.ProgressInfo) {
	output := pr.formatProgressOutput(func() string {
		rateStr := formatTransferSpeed(progress.Rate, pr.formatter.useBytes)
		return fmt.Sprintf("  Download %s", rateStr)
	})
	pr.printOutput(output)
}

// ReportUploadProgress reports upload test progress
func (pr *ProgressReporter) ReportUploadProgress(progress transfer.ProgressInfo) {
	output := pr.formatProgressOutput(func() string {
		rateStr := formatTransferSpeed(progress.Rate, pr.formatter.useBytes)
		return fmt.Sprintf("    Upload %s", rateStr)
	})
	pr.printOutput(output)
}

func (pr *ProgressReporter) formatProgressOutput(activeLine func() string) string {
	var sb strings.Builder

	// Generate output for each state
	states := []types.OutputState{
		types.StatePing,
		types.StateDownload,
		types.StateUpload,
	}

	for _, state := range states {
		line := "  "
		if state == pr.state {
			line = activeLine()
		}

		switch state {
		case types.StatePing:
			sb.WriteString(fmt.Sprintf("      Ping %s\n", line))
		case types.StateDownload:
			sb.WriteString(fmt.Sprintf("  Download %s\n", line))
		case types.StateUpload:
			sb.WriteString(fmt.Sprintf("    Upload %s\n", line))
		}
	}

	return sb.String()
}

func (pr *ProgressReporter) printOutput(output string) {
	// Clear previous output and print new
	// In a real implementation, use ANSI escape codes
	fmt.Print("\r" + output)
	pr.lastOutput = output
}

// Spinner provides a simple spinner animation
type Spinner struct {
	frames   []string
	position int
	interval time.Duration
	lastTick time.Time
}

// NewSpinner creates a new spinner
func NewSpinner() *Spinner {
	return &Spinner{
		frames:   []string{"+", "-", "\\", "|"},
		position: 0,
		interval: 100 * time.Millisecond,
	}
}

// Next returns the next frame
func (s *Spinner) Next() string {
	now := time.Now()
	if now.Sub(s.lastTick) < s.interval {
		return s.frames[s.position]
	}

	s.position = (s.position + 1) % len(s.frames)
	s.lastTick = now
	return s.frames[s.position]
}

func formatTransferSpeed(bytesPerSecond float64, useBytes bool) string {
	if useBytes {
		mbPerSecond := bytesPerSecond / (1024 * 1024)
		return fmt.Sprintf("%.2f MBps", mbPerSecond)
	}
	mbps := bytesPerSecond * 8 / (1000 * 1000)
	return fmt.Sprintf("%.2f Mbps", mbps)
}
