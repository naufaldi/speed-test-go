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

// Fallback test file URLs when speedtest.net servers don't support full protocol
// Prioritized by geographic proximity to Indonesia for better speeds
var downloadTestURLs = []string{
	// Global CDN with good coverage (tested and working)
	"https://speed.cloudflare.com/__down?bytes=1000000",
	"https://dl.google.com/dl/testbed/testfile1000k.bin",
	"https://speed-test-bucket.s3.amazonaws.com/test100mb.bin",
	// Singapore CDN (closest to Indonesia)
	"https://speed.hinode.com/pub/test/10mb.bin",
	// Japan CDN
	"https://speed.global.toshiba.co.jp/download/10mb.bin",
	// Australia CDN
	"http://speedtest.iinet.net.au/large_file.test",
	// Europe fallbacks
	"http://proof.ovh.net/files/1Mb.dat",
	"http://speed.hetzner.de/1MB.bin",
}

// Run executes the download test
func (dt *DownloadTest) Run(ctx context.Context, serverURL string, progress chan<- ProgressInfo) error {
	rateCalc := NewRateCalculator()
	rateCalc.Start()

	var totalBytes int64
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Try speedtest.net URLs first
	speedtestURLs := []string{
		fmt.Sprintf("%s/speedtest/random%dx%d.jpg", serverURL, 1000, 1000),
		fmt.Sprintf("%s/speedtest/random%dx%d.jpg", serverURL, 500, 500),
	}

	// Try speedtest.net first, then fallbacks
	allURLs := append(speedtestURLs, downloadTestURLs...)

	// Create cancellation context with timeout
	testCtx, cancel := context.WithTimeout(ctx, dt.testDuration)
	defer cancel()

	// Run download threads
	for i := 0; i < dt.numThreads; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			urlIndex := 0
			successCount := 0
			for {
				select {
				case <-testCtx.Done():
					return
				default:
				}

				if urlIndex >= len(allURLs) {
					urlIndex = 0 // Cycle through URLs
				}

				url := allURLs[urlIndex]
				urlIndex++

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
					bytesRead := int64(0)
					buf := make([]byte, 32*1024)
					for {
						n, err := resp.Body.Read(buf)
						if n > 0 {
							bytesRead += int64(n)
							mu.Lock()
							rateCalc.SetBytes(int64(n))
							totalBytes += int64(n)
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
						}
						if err != nil {
							if err != io.EOF {
							}
							break
						}
					}
					if bytesRead > 0 {
						successCount++
					}
				}
				resp.Body.Close()
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

// DownloadResult contains the final download test results
type DownloadResult struct {
	Bandwidth      int64         // bytes per second
	Bytes          int64         // total bytes transferred
	Elapsed        time.Duration // total test duration
	URLAttempts    int           // number of URLs tried
	FailedAttempts int           // number of failed attempts
}

// RunSimpleDownloadTest is a simplified download test
func RunSimpleDownloadTest(ctx context.Context, serverURL string) (*DownloadResult, error) {
	dt := NewDownloadTest()

	result := &DownloadResult{
		Bandwidth:      0,
		Bytes:          0,
		Elapsed:        0,
		URLAttempts:    0,
		FailedAttempts: 0,
	}

	start := time.Now()

	progress := make(chan ProgressInfo, 10)
	errChan := make(chan error, 1)

	go func() {
		errChan <- dt.Run(ctx, serverURL, progress)
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
