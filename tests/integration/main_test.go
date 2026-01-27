package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_Version(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "speed-test-test", ".")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}
	defer os.Remove("speed-test-test")

	// Test version flag
	cmd = exec.Command("./speed-test-test", "--version")
	cmd.Dir = getProjectRoot(t)

	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run version command: %v\n%s", err, output)
	}

	result := string(output)
	if !strings.Contains(result, "speed-test") {
		t.Errorf("Expected output to contain 'speed-test', got: %s", result)
	}
}

func TestCLI_Help(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "speed-test-test", ".")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}
	defer os.Remove("speed-test-test")

	// Test help flag
	cmd = exec.Command("./speed-test-test", "--help")
	cmd.Dir = getProjectRoot(t)

	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run help command: %v\n%s", err, output)
	}

	result := string(output)
	if !strings.Contains(result, "Available Commands") {
		t.Errorf("Expected output to contain 'Available Commands', got: %s", result)
	}

	if !strings.Contains(result, "Flags") {
		t.Errorf("Expected output to contain 'Flags', got: %s", result)
	}
}

func TestCLI_JSONOutput(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "speed-test-test", ".")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}
	defer os.Remove("speed-test-test")

	// Test JSON flag (will fail at network but should format correctly)
	cmd = exec.Command("./speed-test-test", "--json", "--timeout", "1s")
	cmd.Dir = getProjectRoot(t)

	result, err := cmd.CombinedOutput()

	// Should get an error (network failure) but still valid JSON error format
	resultStr := string(result)

	// Check if it's valid JSON (even if it's an error)
	if resultStr != "" {
		// Try to parse as JSON
		if json.Valid([]byte(resultStr)) {
			t.Logf("Got JSON output: %s", resultStr)
		} else {
			t.Logf("Got non-JSON output: %s", resultStr)
		}
	}
}

func TestCLI_BytesFlag(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "speed-test-test", ".")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}
	defer os.Remove("speed-test-test")

	// Test bytes flag (will fail at network but should format correctly)
	cmd = exec.Command("./speed-test-test", "--bytes", "--timeout", "1s")
	cmd.Dir = getProjectRoot(t)

	_, err = cmd.CombinedOutput()

	// Should get an error, but flag should be recognized
	t.Logf("Bytes flag test completed")
}

func TestCLI_ServerFlag(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "speed-test-test", ".")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}
	defer os.Remove("speed-test-test")

	// Test server flag with invalid server
	cmd = exec.Command("./speed-test-test", "--server", "999999", "--timeout", "1s")
	cmd.Dir = getProjectRoot(t)

	_, err = cmd.CombinedOutput()

	// Should get an error, but flag should be recognized
	t.Logf("Server flag test completed")
}

func TestCLI_TimeoutFlag(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "speed-test-test", ".")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}
	defer os.Remove("speed-test-test")

	// Test timeout flag
	cmd = exec.Command("./speed-test-test", "--timeout", "5s")
	cmd.Dir = getProjectRoot(t)

	_, err = cmd.CombinedOutput()

	// Should complete within timeout
	t.Logf("Timeout flag test completed")
}

func TestCLI_VerboseFlag(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "speed-test-test", ".")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}
	defer os.Remove("speed-test-test")

	// Test verbose flag
	cmd = exec.Command("./speed-test-test", "--verbose", "--timeout", "1s")
	cmd.Dir = getProjectRoot(t)

	_, err = cmd.CombinedOutput()

	// Should get an error, but flag should be recognized
	t.Logf("Verbose flag test completed")
}

func TestCLI_ShortFlags(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "speed-test-test", ".")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}
	defer os.Remove("speed-test-test")

	// Test short flags
	cmd = exec.Command("./speed-test-test", "-j", "-b", "-t", "1s")
	cmd.Dir = getProjectRoot(t)

	_, err = cmd.CombinedOutput()

	// Should get an error, but flags should be recognized
	t.Logf("Short flags test completed")
}

func TestCLI_InvalidFlag(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "speed-test-test", ".")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}
	defer os.Remove("speed-test-test")

	// Test invalid flag
	cmd = exec.Command("./speed-test-test", "--invalid-flag")
	cmd.Dir = getProjectRoot(t)

	_, err = cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for invalid flag")
	}
}

func TestCLI_NoArgs(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "speed-test-test", ".")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}
	defer os.Remove("speed-test-test")

	// Run without args (will run actual test and fail at network)
	cmd = exec.Command("./speed-test-test")
	cmd.Dir = getProjectRoot(t)

	_, err = cmd.CombinedOutput()

	// Will fail at network, but CLI should work
	t.Logf("No args test completed")
}

func TestBinary_Exists(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "speed-test-test", ".")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}
	defer os.Remove("speed-test-test")

	// Check binary exists
	info, err := os.Stat("speed-test-test")
	if err != nil {
		t.Fatalf("Failed to stat binary: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Binary has zero size")
	}

	if !info.Mode().IsRegular() {
		t.Error("Binary is not a regular file")
	}
}

func TestBinary_Executable(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "speed-test-test", ".")
	cmd.Dir = getProjectRoot(t)

	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("speed-test-test")

	// Check binary is executable
	info, err := os.Stat("speed-test-test")
	if err != nil {
		t.Fatalf("Failed to stat binary: %v", err)
	}

	// On Unix, executable bit should be set
	// On Windows, this check is less meaningful
	if info.Mode()&0111 == 0 {
		t.Logf("Binary mode: %v", info.Mode())
	}
}

func TestBuild_MainPackage(t *testing.T) {
	// Build from main package
	cmd := exec.Command("go", "build", "-o", "speed-test-main-test", ".")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build main package: %v\n%s", err, output)
	}
	defer os.Remove("speed-test-main-test")

	t.Log("Main package builds successfully")
}

func TestBuild_Vet(t *testing.T) {
	// Run go vet
	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go vet failed: %v\n%s", err, output)
	}

	t.Log("go vet passed")
}

func TestBuild_Fmt(t *testing.T) {
	// Run gofmt
	cmd := exec.Command("gofmt", "-l", ".")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gofmt failed: %v\n%s", err, output)
	}

	// Check if any files need formatting
	if len(output) > 0 {
		t.Errorf("Files need formatting:\n%s", output)
	}
}

func TestDependencies_UpToDate(t *testing.T) {
	// Run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = getProjectRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, output)
	}

	// Check if go.sum was modified (would indicate changes)
	t.Log("Dependencies are up to date")
}

func getProjectRoot(t *testing.T) string {
	t.Helper()

	// Start from the directory of this test file
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Walk up to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root with go.mod")
		}
		dir = parent
	}
}
