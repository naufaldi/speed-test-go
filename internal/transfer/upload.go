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
		numThreads:   4,
		testDuration: 10 * time.Second,
		captureFreq:  100 * time.Millisecond,
		uploadSize:   32 * 1024 * 1024, // 32MB per thread
	}
}

// Run executes the upload test
func (ut *UploadTest) Run(ctx context.Context, serverURL string, progress chan<- ProgressInfo) error {
	rateCalc := NewRateCalculator()
	rateCalc.Start()

	var totalBytes int64
	var mu sync.Mutex
	var wg sync.WaitGroup
	progressChan := make(chan ProgressInfo, 10)

	// Start progress reporter
	go func() {
		ticker := time.NewTicker(ut.captureFreq)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case p := <-progressChan:
				totalBytes = p.BytesTotal
				if progress != nil {
					progress <- p
				}
			case <-ticker.C:
				if progress != nil {
					progress <- ProgressInfo{
						Rate:       rateCalc.Rate(),
						BytesTotal: totalBytes,
						Progress:   0.5, // Placeholder
					}
				}
			}
		}
	}()

	// Create cancellation context with timeout
	testCtx, cancel := context.WithTimeout(ctx, ut.testDuration)
	defer cancel()

	// Generate random upload data
	uploadData := make([]byte, ut.uploadSize)
	rand.Read(uploadData)

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

				url := fmt.Sprintf("%s/speedtest/upload.php", serverURL)

				req, err := http.NewRequestWithContext(testCtx, "POST", url, bytes.NewReader(uploadData))
				if err != nil {
					continue
				}
				req.Header.Set("Content-Type", "application/octet-stream")

				resp, err := ut.client.Do(req)
				if err != nil {
					continue
				}

				// Read response to ensure upload completed
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					mu.Lock()
					totalBytes += ut.uploadSize
					rateCalc.SetBytes(ut.uploadSize)
					mu.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()

	// Send final progress
	if progress != nil {
		progress <- ProgressInfo{
			Rate:       rateCalc.Rate(),
			BytesTotal: totalBytes,
			Progress:   1.0,
		}
	}

	return nil
}

// UploadResult contains the final upload test results
type UploadResult struct {
	Bandwidth int64         // bytes per second
	Bytes     int64         // total bytes transferred
	Elapsed   time.Duration // total test duration
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

	select {
	case err := <-errChan:
		result.Elapsed = time.Since(start)
		if err != nil {
			return nil, err
		}
	case <-ctx.Done():
		result.Elapsed = time.Since(start)
		return result, ctx.Err()
	case p := <-progress:
		result.Bytes = p.BytesTotal
		result.Bandwidth = int64(p.Rate)
	}

	return result, nil
}
