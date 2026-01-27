---
name: speed-test-cli-builder
description: Build CLI speed test tools using Go with speedtest.net protocol. Use when: (1) Building a new speed test CLI from scratch, (2) Implementing download/upload bandwidth testing, (3) Creating ping/latency measurement tools, (4) Building server discovery and selection systems, (5) Implementing rate calculation with EWMA, (6) Creating multi-format output formatters (JSON/CSV/human-readable), (7) Adding concurrent transfer testing with goroutines. Reference implementation in this project with full architecture, protocol documentation, and test patterns.
---

# Speed Test CLI Builder Skill

## Quick Reference

**Reference Implementation:** `./`

**Core Flow:** Location Detection → Server Discovery → Ping Test → Download Test → Upload Test

**Protocol Endpoints:**
- Server List: `https://www.speedtest.net/api/js/servers?engine=js&limit=10`
- Ping: `{server}/speedtest/latency.txt`
- Download: `{server}/speedtest/random{size}X{size}.jpg`
- Upload: `{server}/speedtest/upload.php`

## Project Structure

```
.
├── cmd/root.go                    # Cobra CLI framework
├── internal/
│   ├── test/runner.go             # Test orchestration
│   ├── test/ping.go               # Latency measurement
│   ├── transfer/
│   │   ├── download.go            # Download speed test
│   │   ├── upload.go              # Upload speed test
│   │   └── rate.go                # EWMA rate calculation
│   ├── server/discovery.go        # Server list fetching
│   ├── location/client.go         # User location detection
│   └── output/formatter.go        # Multi-format output
└── pkg/types/types.go             # Type definitions
```

## Core Patterns

### 1. Test Runner Orchestration (runner.go:28-99)

```go
func (r *Runner) Run(ctx context.Context) (*types.SpeedTestResult, error) {
    // Step 1: Detect user location
    loc, err := location.DetectUserLocation(ctx)
    
    // Step 2: Fetch and sort servers
    servers, err := server.FetchServerList(ctx)
    server.CalculateServerDistances(servers, userLat, userLon)
    server.SortServersByDistance(servers)
    
    // Step 3: Select best server
    bestServer := servers[0]
    
    // Step 4: Run ping test
    pingResult, err := RunPingTest(ctx, serverURL)
    
    // Step 5: Run download test
    downloadResult, err := transfer.RunSimpleDownloadTest(ctx, serverURL)
    
    // Step 6: Run upload test
    uploadResult, err := transfer.RunSimpleUploadTest(ctx, serverURL)
    
    return result, nil
}
```

### 2. Concurrent Ping Test (ping.go:32-84)

```go
type PingTest struct {
    client      *http.Client
    numPings    int           // 5
    pingTimeout time.Duration // 5s
}

func (pt *PingTest) Run(ctx context.Context, serverURL string) ([]time.Duration, error) {
    latencyURL := fmt.Sprintf("%s/speedtest/latency.txt", serverURL)
    
    sem := make(chan struct{}, 3) // Max 3 concurrent pings
    var latencies []time.Duration
    var mu sync.Mutex
    
    for i := 0; i < pt.numPings; i++ {
        sem <- struct{}{} // Acquire semaphore
        go func(pingNum int) {
            defer wg.Done()
            defer func() { <-sem }() // Release semaphore
            
            start := time.Now()
            req, err := http.NewRequestWithContext(ctx, "GET", latencyURL, nil)
            resp, err := pt.client.Do(req)
            if resp.StatusCode == http.StatusOK {
                latency := time.Since(start)
                mu.Lock()
                latencies = append(latencies, latency)
                mu.Unlock()
            }
        }(i)
    }
    wg.Wait()
    return latencies, nil
}

// Calculate latency and jitter
func CalculateLatency(latencies []time.Duration) (float64, float64) {
    var sum float64
    for _, l := range latencies {
        sum += l.Seconds()
    }
    avg := sum / float64(len(latencies))
    
    var varianceSum float64
    for _, l := range latencies {
        diff := l.Seconds() - avg
        varianceSum += diff * diff
    }
    stdDev := math.Sqrt(varianceSum / float64(len(latencies)))
    
    return avg * 1000, stdDev * 1000 // Convert to milliseconds
}
```

### 3. Download Test (download.go:33-136)

```go
type DownloadTest struct {
    client       *http.Client
    numThreads   int           // 4
    testDuration time.Duration // 15s
    captureFreq  time.Duration // 100ms
}

func (dt *DownloadTest) Run(ctx context.Context, serverURL string, progress chan<- ProgressInfo) error {
    sizes := []int{100, 250, 500, 750, 1000, 1500, 2000, 2500, 3000, 3500}
    rateCalc := NewRateCalculator()
    rateCalc.Start()
    
    var totalBytes int64
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    testCtx, cancel := context.WithTimeout(ctx, dt.testDuration)
    defer cancel()
    
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
                url := fmt.Sprintf("%s/speedtest/random%dX%d.jpg", serverURL, size, size)
                
                req, err := http.NewRequestWithContext(testCtx, "GET", url, nil)
                resp, err := dt.client.Do(req)
                
                buf := make([]byte, 32*1024) // 32KB buffer
                for {
                    n, err := resp.Body.Read(buf)
                    if n > 0 {
                        mu.Lock()
                        rateCalc.SetBytes(int64(n))
                        totalBytes += int64(n)
                        mu.Unlock()
                    }
                    if err != nil {
                        if err != io.EOF { break }
                        break
                    }
                }
                resp.Body.Close()
            }
        }(i)
    }
    wg.Wait()
    return nil
}
```

### 4. Upload Test (upload.go:37-132)

```go
type UploadTest struct {
    client       *http.Client
    numThreads   int           // 4
    testDuration time.Duration // 10s
    uploadSize   int64         // 32MB
}

func (ut *UploadTest) Run(ctx context.Context, serverURL string, progress chan<- ProgressInfo) error {
    rateCalc := NewRateCalculator()
    rateCalc.Start()
    
    uploadData := make([]byte, ut.uploadSize)
    rand.Read(uploadData) // Random data to prevent compression
    
    testCtx, cancel := context.WithTimeout(ctx, ut.testDuration)
    defer cancel()
    
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
                req.Header.Set("Content-Type", "application/octet-stream")
                
                resp, err := ut.client.Do(req)
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
    return nil
}
```

### 5. EWMA Rate Calculation (rate.go:36-80)

```go
type EWMA struct {
    alpha       float64 // 0.1
    value       float64
    initialized bool
}

func (e *EWMA) Update(newValue float64) float64 {
    if !e.initialized {
        e.value = newValue
        e.initialized = true
        return e.value
    }
    e.value = e.alpha*newValue + (1-e.alpha)*e.value
    return e.value
}

type RateCalculator struct {
    ewma        *EWMA
    startTime   time.Time
    startBytes  int64
    currentRate float64
}

func (rc *RateCalculator) SetBytes(bytes int64) {
    delta := bytes - rc.startBytes
    rc.startBytes = bytes
    rc.Update(delta)
}

func (rc *RateCalculator) Rate() float64 {
    return rc.currentRate
}
```

### 6. Server Discovery (discovery.go:17-49)

```go
const serverListURL = "https://www.speedtest.net/api/js/servers?engine=js&limit=10"

func FetchServerList(ctx context.Context) ([]*types.Server, error) {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    req, err := http.NewRequestWithContext(ctx, "GET", serverListURL, nil)
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
    
    client := &http.Client{Timeout: 15 * time.Second}
    resp, err := client.Do(req)
    
    var serverList types.ServerList
    decoder := xml.NewDecoder(resp.Body)
    decoder.Decode(&serverList)
    
    return serverList.Servers, nil
}

// Haversine distance calculation
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
    const earthRadius = 6371.0 // km
    // Convert to radians and apply formula
}
```

### 7. CLI with Cobra (cmd/root.go:21-60)

```go
var (
    jsonFlag     bool
    bytesFlag    bool
    verboseFlag  bool
    serverIDFlag string
    timeoutFlag  time.Duration
)

var rootCmd = &cobra.Command{
    Use:   "speed-test",
    Short: "Test your internet connection speed and ping",
    RunE: runSpeedTest,
}

func init() {
    rootCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output as JSON")
    rootCmd.Flags().BoolVarP(&bytesFlag, "bytes", "b", false, "Output in MB/s")
    rootCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Detailed info")
    rootCmd.Flags().StringVarP(&serverIDFlag, "server", "s", "", "Server ID")
    rootCmd.Flags().DurationVarP(&timeoutFlag, "timeout", "t", 30*time.Second, "Timeout")
}

func runSpeedTest(cmd *cobra.Command, args []string) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeoutFlag)
    defer cancel()
    
    formatter := output.NewFormatter(bytesFlag, jsonFlag, verboseFlag)
    runner := test.NewRunner()
    result, err := runner.Run(ctx)
    
    fmt.Print(formatter.Format(result))
    return nil
}
```

### 8. Multi-Format Output (formatter.go:28-65)

```go
func (f *Formatter) Format(result *types.SpeedTestResult) string {
    if f.useJSON {
        return f.formatJSON(result)
    }
    return f.formatHuman(result)
}

func (f *Formatter) formatHuman(result *types.SpeedTestResult) string {
    var sb strings.Builder
    
    downloadStr := formatSpeed(result.Download.Bandwidth, f.useBytes)
    uploadStr := formatSpeed(result.Upload.Bandwidth, f.useBytes)
    pingStr := fmt.Sprintf("%.1f ms", result.Ping.Latency)
    
    sb.WriteString(fmt.Sprintf("      Ping %s\n", pingStr))
    sb.WriteString(fmt.Sprintf("  Download %s\n", downloadStr))
    sb.WriteString(fmt.Sprintf("    Upload %s\n", uploadStr))
    
    if f.useVerbose && result.Server != nil {
        sb.WriteString(fmt.Sprintf("    Server   %s\n", result.Server.Host))
        sb.WriteString(fmt.Sprintf("  Location   %s (%s)\n", result.Server.Name, result.Server.Country))
    }
    return sb.String()
}

func formatSpeed(bytesPerSecond int64, useBytes bool) string {
    if useBytes {
        mbPerSecond := float64(bytesPerSecond) / (1024 * 1024)
        return fmt.Sprintf("%.2f MBps", mbPerSecond)
    }
    mbps := float64(bytesPerSecond) * 8 / (1000 * 1000)
    return fmt.Sprintf("%.2f Mbps", mbps)
}
```

## Configuration Defaults

| Component | Value |
|-----------|-------|
| numThreads | 4 |
| testDuration (download) | 15s |
| testDuration (upload) | 10s |
| captureFreq | 100ms |
| bufferSize | 32KB |
| uploadSize | 32MB |
| numPings | 5 |
| pingTimeout | 5s |
| EWMA alpha | 0.1 |
| maxConcurrentPings | 3 |

## Key Design Principles

1. **Context-Based Cancellation** - All operations respect context timeout
2. **Concurrent Testing** - Goroutines with WaitGroup for parallel transfers
3. **Rate Limiting** - Semaphore pattern for concurrent ping control
4. **Thread-Safe Updates** - Mutex for shared rate calculation
5. **EWMA Smoothing** - Stable rate readings with 10% new value weight
6. **Multi-Format Output** - JSON, human-readable, verbose modes
7. **XML Protocol** - Parse speedtest.net server list and config

## Testing Checklist

- [ ] Ping returns latency and jitter measurements
- [ ] Download test measures throughput with 4 concurrent threads
- [ ] Upload test POSTs random data to upload.php
- [ ] Rate calculation uses EWMA with alpha=0.1
- [ ] Server discovery fetches and parses XML
- [ ] Distance calculation uses Haversine formula
- [ ] CLI outputs JSON format correctly
- [ ] CLI outputs human-readable format
- [ ] Context timeout cancels all operations
- [ ] Progress channel updates every 100ms

## Common Issues

| Issue | Solution |
|-------|----------|
| No successful pings | Check server URL, increase timeout |
| Slow download | Increase thread count, check buffer size |
| Inconsistent rates | Adjust EWMA alpha value |
| XML parse error | Verify User-Agent header |
| Context timeout | Increase timeoutFlag or reduce test duration |
