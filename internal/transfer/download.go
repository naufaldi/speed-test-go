package transfer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/user/speed-test-go/internal/network"
)

// DownloadTest performs download speed testing
type DownloadTest struct {
	client       *http.Client
	numThreads   int
	testDuration time.Duration
	captureFreq  time.Duration
}

// NewDownloadTest creates a new download test instance
func NewDownloadTest() *DownloadTest {
	return &DownloadTest{
		client:       network.DownloadClient(),
		numThreads:   4,
		testDuration: 15 * time.Second,
		captureFreq:  100 * time.Millisecond,
	}
}

// Run executes the download test
func (dt *DownloadTest) Run(ctx context.Context, serverURL string, progress chan<- ProgressInfo) error {
	rateCalc := NewRateCalculator()
	rateCalc.Start()

	var totalBytes int64
	var mu sync.Mutex
	var wg sync.WaitGroup
	progressChan := make(chan ProgressInfo, 10)

	// Start progress reporter
	go func() {
		ticker := time.NewTicker(dt.captureFreq)
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
						Progress:   0.5,
					}
				}
			}
		}
	}()

	// Create cancellation context with timeout
	testCtx, cancel := context.WithTimeout(ctx, dt.testDuration)
	defer cancel()

	// Try different URL patterns for download test
	urlPatterns := []string{
		"%s/speedtest/random%dx%d.jpg",
		"%s/speedtest/random%dx%d.png",
		"%s/speedtest/garbage.php",
		"%s/speedtest/upload.php",
	}

	sizes := []int{100, 250, 500, 750, 1000, 1500, 2000, 2500, 3000, 3500}

	// Run download threads
	for i := 0; i < dt.numThreads; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			for {
				select {
				case <-testCtx.Done():
					return
				default:
				}

				size := sizes[threadID%len(sizes)]

				success := false
				for _, pattern := range urlPatterns {
					url := fmt.Sprintf(pattern, serverURL, size, size)

					req, err := http.NewRequestWithContext(testCtx, "GET", url, nil)
					if err != nil {
						continue
					}
					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

					resp, err := dt.client.Do(req)
					if err != nil {
						continue
					}

					if resp.StatusCode == http.StatusOK {
						buf := make([]byte, 32*1024)
						for {
							n, err := resp.Body.Read(buf)
							if n > 0 {
								mu.Lock()
								rateCalc.SetBytes(int64(n))
								totalBytes += int64(n)
								mu.Unlock()
								success = true
							}
							if err != nil {
								if err != io.EOF {
								}
								break
							}
						}
					}
					resp.Body.Close()

					if success {
						break
					}
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

// DownloadResult contains the final download test results
type DownloadResult struct {
	Bandwidth int64         // bytes per second
	Bytes     int64         // total bytes transferred
	Elapsed   time.Duration // total test duration
}

// RunSimpleDownloadTest is a simplified download test
func RunSimpleDownloadTest(ctx context.Context, serverURL string) (*DownloadResult, error) {
	dt := NewDownloadTest()

	result := &DownloadResult{
		Bandwidth: 0,
		Bytes:     0,
		Elapsed:   0,
	}

	start := time.Now()

	progress := make(chan ProgressInfo, 10)
	errChan := make(chan error, 1)

	go func() {
		errChan <- dt.Run(ctx, serverURL, progress)
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
