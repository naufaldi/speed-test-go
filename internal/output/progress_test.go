package output

import (
	"strings"
	"testing"
	"time"

	"github.com/user/speed-test-go/internal/transfer"
	"github.com/user/speed-test-go/pkg/types"
)

func TestNewProgressReporter(t *testing.T) {
	f := NewFormatter(false, false, false)
	pr := NewProgressReporter(f)

	if pr == nil {
		t.Fatal("Expected non-nil ProgressReporter")
	}

	if pr.formatter != f {
		t.Error("Expected formatter to be set")
	}

	if pr.updateFreq != 50*time.Millisecond {
		t.Errorf("Expected update frequency 50ms, got: %v", pr.updateFreq)
	}

	if pr.state != types.StateIdle {
		t.Errorf("Expected initial state StateIdle, got: %v", pr.state)
	}
}

func TestProgressReporter_SetState(t *testing.T) {
	f := NewFormatter(false, false, false)
	pr := NewProgressReporter(f)

	// Test state transitions
	states := []types.OutputState{
		types.StateIdle,
		types.StatePing,
		types.StateDownload,
		types.StateUpload,
		types.StateDone,
	}

	for _, state := range states {
		pr.SetState(state)
		if pr.state != state {
			t.Errorf("Expected state %v, got: %v", state, pr.state)
		}
	}
}

func TestProgressReporter_ReportPingProgress(t *testing.T) {
	f := NewFormatter(false, false, false)
	pr := NewProgressReporter(f)

	pr.SetState(types.StatePing)

	// Should not panic
	pr.ReportPingProgress(25.5)
}

func TestProgressReporter_ReportDownloadProgress(t *testing.T) {
	f := NewFormatter(false, false, false)
	pr := NewProgressReporter(f)

	pr.SetState(types.StateDownload)

	progress := transfer.ProgressInfo{
		Rate:       1024 * 1024,
		BytesTotal: 1024 * 1024 * 5,
		Progress:   0.5,
	}

	// Should not panic
	pr.ReportDownloadProgress(progress)
}

func TestProgressReporter_ReportUploadProgress(t *testing.T) {
	f := NewFormatter(false, false, false)
	pr := NewProgressReporter(f)

	pr.SetState(types.StateUpload)

	progress := transfer.ProgressInfo{
		Rate:       1024 * 1024,
		BytesTotal: 1024 * 1024 * 5,
		Progress:   0.5,
	}

	// Should not panic
	pr.ReportUploadProgress(progress)
}

func TestProgressReporter_FormatProgressOutput(t *testing.T) {
	f := NewFormatter(false, false, false)
	pr := NewProgressReporter(f)

	pr.SetState(types.StatePing)

	output := pr.formatProgressOutput(func() string {
		return "      Ping 25.5 ms"
	})

	// Should contain output for all states
	if !strings.Contains(output, "Ping") {
		t.Error("Expected output to contain 'Ping'")
	}

	if !strings.Contains(output, "Download") {
		t.Error("Expected output to contain 'Download'")
	}

	if !strings.Contains(output, "Upload") {
		t.Error("Expected output to contain 'Upload'")
	}

	// The active state should show the actual value
	if !strings.Contains(output, "25.5") {
		t.Error("Expected active state to show ping value")
	}
}

func TestProgressReporter_FormatProgressOutput_Download(t *testing.T) {
	f := NewFormatter(false, false, false)
	pr := NewProgressReporter(f)

	pr.SetState(types.StateDownload)

	output := pr.formatProgressOutput(func() string {
		return "  Download 10.00 Mbps"
	})

	// Download line should be active
	if !strings.Contains(output, "Download 10.00 Mbps") {
		t.Error("Expected active download line with value")
	}
}

func TestProgressReporter_FormatProgressOutput_Upload(t *testing.T) {
	f := NewFormatter(false, false, false)
	pr := NewProgressReporter(f)

	pr.SetState(types.StateUpload)

	output := pr.formatProgressOutput(func() string {
		return "    Upload 5.00 Mbps"
	})

	// Upload line should be active
	if !strings.Contains(output, "Upload 5.00 Mbps") {
		t.Error("Expected active upload line with value")
	}
}

func TestProgressReporter_PrintOutput(t *testing.T) {
	f := NewFormatter(false, false, false)
	pr := NewProgressReporter(f)

	// Should not panic
	pr.printOutput("test output")

	if pr.lastOutput != "test output" {
		t.Errorf("Expected lastOutput to be 'test output', got: %s", pr.lastOutput)
	}
}

func TestNewSpinner(t *testing.T) {
	s := NewSpinner()

	if s == nil {
		t.Fatal("Expected non-nil Spinner")
	}

	if len(s.frames) != 4 {
		t.Errorf("Expected 4 frames, got: %d", len(s.frames))
	}

	if s.position != 0 {
		t.Errorf("Expected initial position 0, got: %d", s.position)
	}

	if s.interval != 100*time.Millisecond {
		t.Errorf("Expected interval 100ms, got: %v", s.interval)
	}
}

func TestSpinner_Next(t *testing.T) {
	s := NewSpinner()

	// Get first frame
	frame1 := s.Next()
	if frame1 == "" {
		t.Error("Expected non-empty frame")
	}

	// Should return same frame if interval hasn't passed
	frame2 := s.Next()
	if frame1 != frame2 {
		t.Error("Expected same frame when interval hasn't passed")
	}

	// Advance time
	time.Sleep(150 * time.Millisecond)

	// Should return next frame
	frame3 := s.Next()
	if frame1 == frame3 {
		t.Error("Expected different frame after interval")
	}
}

func TestSpinner_WrapAround(t *testing.T) {
	s := NewSpinner()

	// Get all frames
	frames := make(map[string]bool)
	for i := 0; i < 10; i++ {
		frames[s.Next()] = true
		time.Sleep(150 * time.Millisecond)
	}

	if len(frames) != 4 {
		t.Errorf("Expected 4 unique frames, got: %d", len(frames))
	}
}

func TestSpinner_DefaultFrame(t *testing.T) {
	s := NewSpinner()

	// Initial call should return first frame
	frame := s.Next()

	if frame == "" {
		t.Error("Expected non-empty frame")
	}

	// The first frame should be one of the valid frames
	validFrames := map[string]bool{"+": true, "-": true, "\\": true, "|": true}
	if !validFrames[frame] {
		t.Errorf("Expected valid frame (+, -, \\, |), got: %s", frame)
	}
}

func TestFormatTransferSpeed_Mbps(t *testing.T) {
	// 1,250,000 bytes = 10 Mbps
	result := formatTransferSpeed(1250000, false)

	if result != "10.00 Mbps" {
		t.Errorf("Expected '10.00 Mbps', got: %s", result)
	}
}

func TestFormatTransferSpeed_MBps(t *testing.T) {
	result := formatTransferSpeed(1048576, true) // 1 MB/s

	if result != "1.00 MBps" {
		t.Errorf("Expected '1.00 MBps', got: %s", result)
	}
}

func TestFormatTransferSpeed_Zero(t *testing.T) {
	result := formatTransferSpeed(0, false)

	if result != "0.00 Mbps" {
		t.Errorf("Expected '0.00 Mbps', got: %s", result)
	}
}

func TestFormatTransferSpeed_SmallValue(t *testing.T) {
	result := formatTransferSpeed(1000, false) // Very small value

	if !strings.Contains(result, "Mbps") {
		t.Error("Expected 'Mbps' in result")
	}
}

func TestFormatTransferSpeed_LargeValue(t *testing.T) {
	// 125,000,000 bytes = 1000 Mbps
	result := formatTransferSpeed(125000000, false)

	if !strings.Contains(result, "Mbps") {
		t.Error("Expected 'Mbps' in result")
	}

	// Should handle large values (1000.00 Mbps)
	if !strings.Contains(result, "1000.00") {
		t.Errorf("Expected '1000.00 Mbps' for 1000 Mbps, got: %s", result)
	}
}

func TestProgressReporter_StateMachine(t *testing.T) {
	f := NewFormatter(false, false, false)
	pr := NewProgressReporter(f)

	// Test state machine transitions
	testCases := []struct {
		initialState types.OutputState
		newState     types.OutputState
	}{
		{types.StateIdle, types.StatePing},
		{types.StatePing, types.StateDownload},
		{types.StateDownload, types.StateUpload},
		{types.StateUpload, types.StateDone},
	}

	for _, tc := range testCases {
		pr.SetState(tc.initialState)
		pr.SetState(tc.newState)

		if pr.state != tc.newState {
			t.Errorf("Expected state %v, got: %v", tc.newState, pr.state)
		}
	}
}

func TestProgressReporter_ProgressWithBytesMode(t *testing.T) {
	f := NewFormatter(true, false, false)
	pr := NewProgressReporter(f)

	pr.SetState(types.StateDownload)

	progress := transfer.ProgressInfo{
		Rate:       1048576, // 1 MB/s
		BytesTotal: 5242880, // 5 MB
		Progress:   0.5,
	}

	// Should format correctly in bytes mode
	pr.ReportDownloadProgress(progress)
}

func TestProgressReporter_VerboseMode(t *testing.T) {
	f := NewFormatter(false, false, true)
	pr := NewProgressReporter(f)

	pr.SetState(types.StatePing)

	// Should work with verbose formatter
	pr.ReportPingProgress(25.5)
}

func TestProgressReporter_AllStates(t *testing.T) {
	f := NewFormatter(false, false, false)
	pr := NewProgressReporter(f)

	// Test that all states can be set and reported
	states := []types.OutputState{
		types.StateIdle,
		types.StatePing,
		types.StateDownload,
		types.StateUpload,
		types.StateDone,
	}

	for _, state := range states {
		pr.SetState(state)
		if pr.state != state {
			t.Errorf("Failed to set state %v: got %v", state, pr.state)
		}
	}
}

func TestProgressReporter_ConcurrentAccess(t *testing.T) {
	f := NewFormatter(false, false, false)
	pr := NewProgressReporter(f)

	// Test concurrent access
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			pr.SetState(types.StatePing)
			pr.ReportPingProgress(25.5)
			pr.SetState(types.StateDownload)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSpinner_FrameSequence(t *testing.T) {
	// Create a completely fresh spinner with known state
	s := &Spinner{
		frames:    []string{"+", "-", "\\", "|"},
		position:  0,
		interval:  100 * time.Millisecond,
		lastTick:  time.Time{}, // Unix epoch - ensures first call increments
	}

	// Spinner increments position FIRST, then returns the new frame
	// So starting at position 0:
	// - First call: increment to 1, return frames[1] = "-"
	// - Second call: increment to 2, return frames[2] = "\"
	// - Third call: increment to 3, return frames[3] = "|"
	// - Fourth call: increment to 0 (wrap), return frames[0] = "+"
	expectedFrames := []string{"-", "\\", "|", "+"}

	for i, expected := range expectedFrames {
		frame := s.Next()
		if frame != expected {
			t.Errorf("Frame %d: expected '%s', got: '%s'", i, expected, frame)
		}
		// Wait for next frame to be available
		if i < len(expectedFrames)-1 {
			time.Sleep(s.interval + 10*time.Millisecond)
		}
	}

	// Wait for wrap around
	time.Sleep(s.interval + 10*time.Millisecond)

	// Should return "-" again (position 1)
	frame := s.Next()
	if frame != "-" {
		t.Errorf("Expected '-', got: '%s'", frame)
	}
}

func TestProgressReporter_MultipleProgressReports(t *testing.T) {
	f := NewFormatter(false, false, false)
	pr := NewProgressReporter(f)

	pr.SetState(types.StateDownload)

	// Report multiple progress updates
	for i := 0; i < 5; i++ {
		progress := transfer.ProgressInfo{
			Rate:       float64(i) * 1024 * 1024,
			BytesTotal: int64(i) * 1024 * 1024 * 10,
			Progress:   float64(i) / 5.0,
		}
		pr.ReportDownloadProgress(progress)
	}
}
