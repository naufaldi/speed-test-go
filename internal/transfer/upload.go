package transfer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/user/speed-test-go/internal/network"
)

// UploadTest performs upload speed testing
type UploadTest struct {
	client       *http.Client
	numThreads   int
	testDuration time.Duration
	captureFreq  time.Duration
	uploadSize   int64
}

// NewUploadTest creates a new upload test instance
func NewUploadTest() *UploadTest {
	return &UploadTest{
		client:       network.UploadClient(),
		numThreads:   2,
		testDuration: 20 * time.Second,
		captureFreq:  100 * time.Millisecond,
		uploadSize:   1 * 1024 * 1024, // 1MB per thread
	}
}

// Run executes the upload test
func (ut *UploadTest) Run(ctx context.Context, serverURL string, progress chan<- ProgressInfo) error {
	rateCalc := NewRateCalculator()
	rateCalc.Start()

	var totalBytes int64
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Try speedtest.net URLs first, then public echo servers
	// Prioritized by geographic proximity to Indonesia
	uploadURLs := []string{
		fmt.Sprintf("%s/speedtest/upload.php", serverURL),
		fmt.Sprintf("%s/upload.php", serverURL),
		// Asia Pacific echo servers (better latency from Indonesia)
		"https://postman-echo.com/post",
		"https://reqres.in/api/posts",
		// US fallback
		"https://httpbin.org/post",
	}

	// Generate random upload data
	uploadData := make([]byte, ut.uploadSize)
	rand.Read(uploadData)

	// Create cancellation context with timeout
	testCtx, cancel := context.WithTimeout(ctx, ut.testDuration)
	defer cancel()

	// Run upload threads
	for i := 0; i < ut.numThreads; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			for {
				select {
				case <-testCtx.Done():
					return
				default:
				}

				success := false
				for _, url := range uploadURLs {
					req, err := http.NewRequestWithContext(testCtx, "POST", url, bytes.NewReader(uploadData))
					if err != nil {
						continue
					}
					req.Header.Set("Content-Type", "application/octet-stream")
					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

					resp, err := ut.client.Do(req)
					if err != nil {
						continue
					}

					// Read response to complete the request
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					if resp.StatusCode == http.StatusOK {
						mu.Lock()
						rateCalc.SetBytes(ut.uploadSize)
						totalBytes += ut.uploadSize
						mu.Unlock()

						// Send progress update
						if progress != nil {
							select {
							case progress <- ProgressInfo{
								Rate:       rateCalc.Rate(),
								BytesTotal: totalBytes,
								Progress:   0.5,
							}:
							case <-testCtx.Done():
							case <-ctx.Done():
							}
						}
						success = true
						break
					}
				}

				if success {
					break
				}
			}
		}(i)
	}

	wg.Wait()

	// Send final progress
	if progress != nil {
		close(progress)
	}

	return nil
}

// UploadResult contains the final upload test results
type UploadResult struct {
	Bandwidth      int64         // bytes per second
	Bytes          int64         // total bytes transferred
	Elapsed        time.Duration // total test duration
	URLAttempts    int           // number of URLs tried
	FailedAttempts int           // number of failed attempts
}

// RunSimpleUploadTest is a simplified upload test
func RunSimpleUploadTest(ctx context.Context, serverURL string) (*UploadResult, error) {
	ut := NewUploadTest()

	result := &UploadResult{
		Bandwidth: 0,
		Bytes:     0,
		Elapsed:   0,
	}

	start := time.Now()

	progress := make(chan ProgressInfo, 10)
	errChan := make(chan error, 1)

	go func() {
		errChan <- ut.Run(ctx, serverURL, progress)
	}()

	// Wait for completion or timeout
	select {
	case err := <-errChan:
		result.Elapsed = time.Since(start)
		// Get last progress
		for p := range progress {
			result.Bytes = p.BytesTotal
			result.Bandwidth = int64(p.Rate)
		}
		if err != nil {
			return nil, err
		}
	case <-ctx.Done():
		result.Elapsed = time.Since(start)
		for p := range progress {
			result.Bytes = p.BytesTotal
			result.Bandwidth = int64(p.Rate)
		}
		return result, ctx.Err()
	}

	return result, nil
}
