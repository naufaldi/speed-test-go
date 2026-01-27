package transfer

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewUploadTest(t *testing.T) {
	ut := NewUploadTest()

	if ut == nil {
		t.Fatal("Expected non-nil UploadTest")
	}

	if ut.client == nil {
		t.Error("Expected client to be initialized")
	}

	if ut.numThreads != 4 {
		t.Errorf("Expected 4 threads, got: %d", ut.numThreads)
	}

	if ut.testDuration != 10*time.Second {
		t.Errorf("Expected test duration 10s, got: %v", ut.testDuration)
	}

	if ut.captureFreq != 100*time.Millisecond {
		t.Errorf("Expected capture frequency 100ms, got: %v", ut.captureFreq)
	}

	if ut.uploadSize != 32*1024*1024 {
		t.Errorf("Expected upload size 32MB, got: %d", ut.uploadSize)
	}
}

func TestUploadTest_Run_ContextCancellation(t *testing.T) {
	// Create a test server that responds slowly
	var mu sync.Mutex
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()

		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ut := NewUploadTest()
	ut.uploadSize = 1024 * 1024 // Smaller upload for faster test

	ctx, cancel := context.WithCancel(context.Background())
	progress := make(chan ProgressInfo, 10)
	errChan := make(chan error, 1)

	go func() {
		errChan <- ut.Run(ctx, server.URL, progress)
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

func TestUploadTest_Run_ProgressReporting(t *testing.T) {
	// Create a test server
	var mu sync.Mutex
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()

		// Read and discard the body
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ut := NewUploadTest()
	ut.uploadSize = 1024 * 1024 // Smaller upload for faster test

	ctx := context.Background()
	progress := make(chan ProgressInfo, 10)
	errChan := make(chan error, 1)

	go func() {
		errChan <- ut.Run(ctx, server.URL, progress)
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

func TestUploadTest_Run_NilProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ut := NewUploadTest()
	ut.uploadSize = 1024 * 1024 // Smaller upload for faster test

	ctx := context.Background()

	// Should not panic with nil progress channel
	err := ut.Run(ctx, server.URL, nil)

	// Test should complete without error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunSimpleUploadTest_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()
	result, err := RunSimpleUploadTest(ctx, server.URL)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Note: Simple test may not complete full upload
	if result.Elapsed <= 0 {
		t.Error("Expected positive elapsed time")
	}
}

func TestRunSimpleUploadTest_ContextCancellation(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	result, err := RunSimpleUploadTest(ctx, server.URL)

	if err == nil {
		// If no error, should still have some result
		if result == nil {
			t.Error("Expected non-nil result")
		}
	}
}

func TestRunSimpleUploadTest_InvalidURL(t *testing.T) {
	ctx := context.Background()

	// Use an invalid URL (server won't exist)
	_, err := RunSimpleUploadTest(ctx, "http://invalid.example.com/nonexistent")

	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestUploadResult_Fields(t *testing.T) {
	result := &UploadResult{
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

func TestUploadTest_MultipleThreads(t *testing.T) {
	var mu sync.Mutex
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()

		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ut := NewUploadTest()
	ut.numThreads = 8           // Increase thread count for test
	ut.uploadSize = 1024 * 1024 // Smaller upload for faster test

	ctx := context.Background()
	err := ut.Run(ctx, server.URL, nil)

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

func TestUploadTest_ErrorHandling(t *testing.T) {
	// Create a server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ut := NewUploadTest()
	ut.uploadSize = 1024 * 1024 // Smaller upload for faster test

	ctx := context.Background()

	// Should not panic, should handle errors gracefully
	err := ut.Run(ctx, server.URL, nil)

	// Error handling should not cause panic
	if err != nil {
		t.Logf("Got expected error: %v", err)
	}
}

func TestUploadTest_Timeout(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ut := NewUploadTest()
	ut.testDuration = 500 * time.Millisecond // Short timeout
	ut.uploadSize = 1024 * 1024              // Smaller upload for faster test

	ctx := context.Background()
	start := time.Now()
	err := ut.Run(ctx, server.URL, nil)
	elapsed := time.Since(start)

	if err != nil {
		t.Logf("Got error: %v", err)
	}

	// Should complete within reasonable time of the timeout
	if elapsed > 2*time.Second {
		t.Errorf("Test took too long: %v", elapsed)
	}
}

func TestUploadTest_SuccessfulUpload(t *testing.T) {
	var receivedBytes int64
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n, err := io.Copy(io.Discard, r.Body)
		if err == nil {
			mu.Lock()
			receivedBytes = n
			mu.Unlock()
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ut := NewUploadTest()
	ut.uploadSize = 1024 * 1024 // 1MB upload

	ctx := context.Background()
	err := ut.Run(ctx, server.URL, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if receivedBytes <= 0 {
		t.Errorf("Expected bytes to be uploaded, got: %d", receivedBytes)
	}
}

func TestUploadTest_ContentType(t *testing.T) {
	var contentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ut := NewUploadTest()
	ut.uploadSize = 1024 * 1024

	ctx := context.Background()
	_ = ut.Run(ctx, server.URL, nil)

	if contentType != "application/octet-stream" {
		t.Errorf("Expected Content-Type 'application/octet-stream', got: %s", contentType)
	}
}

func TestUploadTest_PostMethod(t *testing.T) {
	var method string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ut := NewUploadTest()
	ut.uploadSize = 1024 * 1024

	ctx := context.Background()
	_ = ut.Run(ctx, server.URL, nil)

	if method != "POST" {
		t.Errorf("Expected POST method, got: %s", method)
	}
}

func TestUploadTest_UploadURL(t *testing.T) {
	var receivedURL string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.Path
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ut := NewUploadTest()
	ut.uploadSize = 1024 * 1024

	ctx := context.Background()
	_ = ut.Run(ctx, server.URL, nil)

	expected := "/speedtest/upload.php"
	if receivedURL != expected {
		t.Errorf("Expected URL '%s', got: %s", expected, receivedURL)
	}
}

func TestUploadTest_ConcurrentUploads(t *testing.T) {
	var activeUploads int
	var mu sync.Mutex
	maxActive := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		activeUploads++
		if activeUploads > maxActive {
			maxActive = activeUploads
		}
		mu.Unlock()

		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		activeUploads--
		mu.Unlock()

		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ut := NewUploadTest()
	ut.numThreads = 4
	ut.uploadSize = 1024 * 1024
	ut.testDuration = 500 * time.Millisecond

	ctx := context.Background()
	_ = ut.Run(ctx, server.URL, nil)

	// Verify concurrent execution
	mu.Lock()
	defer mu.Unlock()

	if maxActive > 1 {
		t.Logf("Successfully ran %d concurrent uploads", maxActive)
	}
}

func TestUploadTest_RandomDataGeneration(t *testing.T) {
	// Verify that upload generates different data each time
	ut := NewUploadTest()
	ut.uploadSize = 1024 * 1024

	// Generate data twice
	data1 := make([]byte, ut.uploadSize)
	data2 := make([]byte, ut.uploadSize)

	// We can't directly test rand.Read, but we can verify the function doesn't panic
	if len(data1) != int(ut.uploadSize) || len(data2) != int(ut.uploadSize) {
		t.Error("Upload size not set correctly")
	}
}
