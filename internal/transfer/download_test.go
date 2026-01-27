package transfer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewDownloadTest(t *testing.T) {
	dt := NewDownloadTest()

	if dt == nil {
		t.Fatal("Expected non-nil DownloadTest")
	}

	if dt.client == nil {
		t.Error("Expected client to be initialized")
	}

	if dt.numThreads != 4 {
		t.Errorf("Expected 4 threads, got: %d", dt.numThreads)
	}

	if dt.testDuration != 15*time.Second {
		t.Errorf("Expected test duration 15s, got: %v", dt.testDuration)
	}

	if dt.captureFreq != 100*time.Millisecond {
		t.Errorf("Expected capture frequency 100ms, got: %v", dt.captureFreq)
	}
}

func TestDownloadTest_Run_ContextCancellation(t *testing.T) {
	// Create a test server that responds slowly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test data"))
	}))
	defer server.Close()

	dt := NewDownloadTest()
	ctx, cancel := context.WithCancel(context.Background())

	progress := make(chan ProgressInfo, 10)
	errChan := make(chan error, 1)

	go func() {
		errChan <- dt.Run(ctx, server.URL, progress)
	}()

	// Cancel after a short delay
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errChan:
		// Should complete after cancellation
		if err != nil {
			t.Logf("Got error (expected for cancellation): %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Test did not complete after cancellation")
	}
}

func TestDownloadTest_Run_ProgressReporting(t *testing.T) {
	// Create a test server
	var mu sync.Mutex
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(make([]byte, 32*1024))) // 32KB response
	}))
	defer server.Close()

	dt := NewDownloadTest()
	ctx := context.Background()

	progress := make(chan ProgressInfo, 10)
	errChan := make(chan error, 1)

	go func() {
		errChan <- dt.Run(ctx, server.URL, progress)
	}()

	// Collect progress updates
	progressUpdates := 0
	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-progress:
			progressUpdates++
		case err := <-errChan:
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if progressUpdates == 0 {
				t.Error("Expected at least one progress update")
			}
			return
		case <-timeout:
			t.Error("Timeout waiting for test completion")
			return
		}
	}
}

func TestDownloadTest_Run_NilProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test data"))
	}))
	defer server.Close()

	dt := NewDownloadTest()
	ctx := context.Background()

	// Should not panic with nil progress channel
	err := dt.Run(ctx, server.URL, nil)

	// Test should complete without error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunSimpleDownloadTest_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(make([]byte, 1024))) // 1KB response
	}))
	defer server.Close()

	ctx := context.Background()
	result, err := RunSimpleDownloadTest(ctx, server.URL)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Bytes <= 0 {
		t.Errorf("Expected positive bytes, got: %d", result.Bytes)
	}

	if result.Elapsed <= 0 {
		t.Error("Expected positive elapsed time")
	}
}

func TestRunSimpleDownloadTest_ContextCancellation(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	result, err := RunSimpleDownloadTest(ctx, server.URL)

	if err == nil {
		// If no error, should still have some result
		if result == nil {
			t.Error("Expected non-nil result")
		}
	}
}

func TestRunSimpleDownloadTest_InvalidURL(t *testing.T) {
	ctx := context.Background()

	// Use an invalid URL (server won't exist)
	_, err := RunSimpleDownloadTest(ctx, "http://invalid.example.com/nonexistent")

	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestDownloadResult_Fields(t *testing.T) {
	result := &DownloadResult{
		Bandwidth: 1024 * 1024,      // 1 MB/s
		Bytes:     1024 * 1024 * 10, // 10 MB
		Elapsed:   10 * time.Second,
	}

	if result.Bandwidth != 1024*1024 {
		t.Errorf("Expected Bandwidth %d, got: %d", 1024*1024, result.Bandwidth)
	}

	if result.Bytes != 1024*1024*10 {
		t.Errorf("Expected Bytes %d, got: %d", 1024*1024*10, result.Bytes)
	}

	if result.Elapsed != 10*time.Second {
		t.Errorf("Expected Elapsed %v, got: %v", 10*time.Second, result.Elapsed)
	}
}

func TestProgressInfo_Fields(t *testing.T) {
	progress := ProgressInfo{
		Rate:       1024 * 1024,     // 1 MB/s
		BytesTotal: 1024 * 1024 * 5, // 5 MB
		Progress:   0.5,
	}

	if progress.Rate != float64(1024*1024) {
		t.Errorf("Expected Rate %d, got: %f", 1024*1024, progress.Rate)
	}

	if progress.BytesTotal != 1024*1024*5 {
		t.Errorf("Expected BytesTotal %d, got: %d", 1024*1024*5, progress.BytesTotal)
	}

	if progress.Progress != 0.5 {
		t.Errorf("Expected Progress 0.5, got: %f", progress.Progress)
	}
}

func TestDownloadTest_MultipleThreads(t *testing.T) {
	var mu sync.Mutex
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(make([]byte, 1024)))
	}))
	defer server.Close()

	dt := NewDownloadTest()
	dt.numThreads = 8 // Increase thread count for test

	ctx := context.Background()
	err := dt.Run(ctx, server.URL, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	// Should have multiple requests (one per thread)
	if requestCount < 1 {
		t.Errorf("Expected at least 1 request, got: %d", requestCount)
	}
}

func TestDownloadTest_ErrorHandling(t *testing.T) {
	// Create a server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	dt := NewDownloadTest()
	ctx := context.Background()

	// Should not panic, should handle errors gracefully
	err := dt.Run(ctx, server.URL, nil)

	// Error handling should not cause panic
	if err != nil {
		t.Logf("Got expected error: %v", err)
	}
}

func TestDownloadTest_Timeout(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dt := NewDownloadTest()
	dt.testDuration = 500 * time.Millisecond // Short timeout

	ctx := context.Background()
	start := time.Now()
	err := dt.Run(ctx, server.URL, nil)
	elapsed := time.Since(start)

	if err != nil {
		t.Logf("Got error: %v", err)
	}

	// Should complete within reasonable time of the timeout
	if elapsed > 2*time.Second {
		t.Errorf("Test took too long: %v", elapsed)
	}
}
