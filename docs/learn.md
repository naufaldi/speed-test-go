# Speed Test CLI - Learning Guide

This guide helps developers understand the speedtest.net protocol, Go networking concepts, and how this implementation works.

## 1. Understanding the Speedtest.net Protocol

### 1.1 Protocol Overview

Speedtest.net uses a multi-phase testing approach to measure network performance:

1. **Server Discovery Phase**: The client fetches a list of available test servers from speedtest.net
2. **User Location Phase**: The client detects the user's geographic location and ISP
3. **Server Selection Phase**: The client selects the optimal server (usually the closest)
4. **Ping Test Phase**: Measures latency to the selected server
5. **Download Test Phase**: Measures download bandwidth
6. **Upload Test Phase**: Measures upload bandwidth

### 1.2 Server Discovery

The speedtest.net API provides a list of test servers via an XML endpoint:

```
GET https://www.speedtest.net/api/js/servers?engine=js&limit=10
```

**XML Response Structure**:
```xml
<servers>
  <server url="http://server1.speedtest.net/speedtest/upload.php"
          lat="40.7128"
          lon="-74.0060"
          name="New York"
          country="United States"
          sponsor="Test ISP"
          id="1234"
          host="server1.speedtest.net:8080"/>
</servers>
```

**Key Fields**:
- `url`: Full URL to the server's upload endpoint
- `lat`, `lon`: Geographic coordinates
- `name`: Server name/location
- `country`: Country code
- `sponsor`: ISP or organization sponsoring the server
- `id`: Unique server identifier
- `host`: Server hostname and port

**Server Selection Algorithm**:
1. Parse all server coordinates
2. Calculate distance from user to each server (Haversine formula)
3. Sort servers by distance (closest first)
4. Ping top N servers to measure actual latency (configurable via `--servers` flag, default 5)
5. Select server with lowest latency

### 1.7 Server Selection

The server selection is a critical component that ensures optimal speed test results. The implementation in `internal/server/selection.go` provides:

**Key Functions:**

```go
// SelectBestServerByPing selects the best server by pinging the top N closest servers
func SelectBestServerByPing(ctx context.Context, servers []*Server, numServers int) (*Server, error)

// FindServerByID searches for a server by its ID
func FindServerByID(servers []*Server, id string) *Server

// GetLatencyResult returns a PingResult for a single server
func GetLatencyResult(ctx context.Context, server *Server) (*PingResult, error)
```

**Selection Process:**

1. **Distance-Based Pre-selection**: Servers are sorted by distance from the user
2. **Latency Testing**: The top N servers are pinged concurrently (max 3 at a time)
3. **Latency Measurement**: Each server is pinged 3 times and averaged
4. **Best Server Selection**: Server with lowest average latency is selected
5. **Fallback**: If all pings fail, closest server is used

**Configuration:**

- `--servers` flag (default: 5): Number of closest servers to test
- `--server` flag: Use a specific server ID instead of auto-selection

**Example:**
```bash
# Test 10 closest servers for selection
speed-test --servers 10

# Use specific server
speed-test --server 1234
```

### 1.3 User Location Detection

The client can detect the user's location by making a request to:

```
GET http://speedtest.net/speedtest-config.php
```

**XML Response Structure**:
```xml
<client ip="203.0.113.42"
        lat="40.7128"
        lon="-74.0060"
        isp="Example ISP"/>
```

**Key Fields**:
- `ip`: Public IP address
- `lat`, `lon`: Detected geographic coordinates
- `isp`: Internet Service Provider name

### 1.4 Ping Test

The ping test measures latency by making HTTP requests to the server's latency endpoint:

```
GET http://server-host/speedtest/latency.txt
```

**Implementation Details**:
- Send multiple ping requests (typically 5-10)
- Use concurrent requests with rate limiting (max 3 concurrent)
- Calculate average latency and jitter (standard deviation)

**Latency Calculation**:
```
Average Latency = (sum of all ping times) / (number of pings)
Jitter = standard deviation of ping times
```

### 1.5 Download Test

The download test measures bandwidth by downloading test files:

```
GET http://server-host/speedtest/random{size}x{size}.jpg
```

**File Sizes**: 100, 250, 500, 750, 1000, 1500, 2000, 2500, 3000, 3500 KB

**Implementation Details**:
- Use multiple concurrent connections (typically 4-8)
- Download files in chunks to measure rate
- Calculate bandwidth using EWMA for smoothing

**Bandwidth Calculation**:
```
Raw Rate = bytes_transferred / time_elapsed
Smoothed Rate = alpha * raw_rate + (1 - alpha) * previous_rate
```

Where alpha is typically 0.1 (10% weight for new values).

### 1.6 Upload Test

The upload test measures bandwidth by POSTing data to the server:

```
POST http://server-host/speedtest/upload.php
Content-Type: application/octet-stream
```

**Implementation Details**:
- Generate random data for upload (typically 32MB per thread)
- Use multiple concurrent uploads
- Measure upload rate

## 2. Go Concepts Used

### 2.1 Concurrency

This project uses Go's concurrency primitives extensively:

**Goroutines**:
```go
go func() {
    // This runs concurrently
}()
```

**Channels**:
```go
// Progress reporting channel
progress := make(chan ProgressInfo, 10)
progress <- ProgressInfo{Rate: 1000000}
```

**Context**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
// Context automatically cancels after timeout
```

**WaitGroup**:
```go
var wg sync.WaitGroup
for i := 0; i < 4; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // Work
    }()
}
wg.Wait() // Wait for all goroutines
```

### 2.2 Networking

**HTTP Client Configuration**:
```go
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout: 10 * time.Second,
        }).DialContext,
    },
}
```

**Request with Context**:
```go
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)
```

**XML Parsing**:
```go
decoder := xml.NewDecoder(resp.Body)
var serverList ServerList
if err := decoder.Decode(&serverList); err != nil {
    return nil, err
}
```

### 2.3 Data Handling

**Struct Tags for JSON/XML**:
```go
type Server struct {
    URL      string  `xml:"url,attr"`
    Lat      string  `xml:"lat,attr"`
    Name     string  `xml:"name,attr"`
    Country  string  `xml:"country,attr"`
    ID       string  `xml:"id,attr"`
}
```

### 2.4 CLI Development with Cobra

**Basic Command Structure**:
```go
var rootCmd = &cobra.Command{
    Use:   "speed-test",
    Short: "Test your internet connection speed and ping",
    RunE:  runSpeedTest,
}
```

**Flag Definition**:
```go
func init() {
    rootCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output as JSON")
    rootCmd.Flags().StringVarP(&serverIDFlag, "server", "s", "", "Server ID")
}
```

## 3. Code Architecture

### 3.1 Project Structure

```
speed-test-go/
├── cmd/
│   ├── root.go           # CLI entry point
│   └── version.go        # Version command
├── internal/
│   ├── location/         # User location detection
│   ├── network/          # HTTP client configuration
│   ├── output/           # Output formatting and progress
│   ├── server/           # Server discovery and selection
│   ├── test/             # Test orchestration
│   └── transfer/         # Download/upload implementation
├── pkg/
│   └── types/            # Public type definitions
├── .github/
│   └── workflows/        # CI/CD configuration
├── docs/                 # Documentation
├── main.go               # Application entry point
├── Makefile              # Build automation
└── go.mod                # Go module definition
```

### 3.2 Component Relationships

```
main.go
    |
    v
cmd/root.go (Cobra CLI)
    |
    v
internal/test/runner.go (Test Orchestration)
    |
    +---> internal/server/discovery.go (Server List)
    +---> internal/server/selection.go (Server Selection)  <-- NEW
    +---> internal/location/client.go (User Location)
    +---> internal/test/ping.go (Ping Test)
    +---> internal/transfer/download.go (Download Test)
    +---> internal/transfer/upload.go (Upload Test)
    |
    v
internal/output/ (Output Formatting)
```

### 3.3 Key Data Types

**SpeedTestResult**: The final output containing all test results
```go
type SpeedTestResult struct {
    Timestamp time.Time        // When test completed
    Ping      PingResult       // Latency measurements
    Download  TransferResult   // Download speed
    Upload    TransferResult   // Upload speed
    Server    *ServerInfo      // Selected server
    Interface *InterfaceInfo   // Network interface
    ISP       string           // Internet provider
}
```

**Server**: Represents a test server from the API
```go
type Server struct {
    URL      string  // Upload endpoint URL
    Lat      string  // Latitude
    Lon      string  // Longitude
    Name     string  // Server name
    Country  string  // Country
    Sponsor  string  // ISP sponsor
    ID       string  // Server ID
    Host     string  // Hostname
    Distance float64 // Distance from user
}
```

## 4. Reference Materials

### 4.1 Primary Reference: sindresorhus/speed-test

The [sindresorhus/speed-test](https://github.com/sindresorhus/speed-test) Node.js CLI is the primary reference for this project.

**What to Study**:
- CLI flag interface (`--json`, `--bytes`, `--verbose`)
- Output format and layout (Ping, Download, Upload)
- Progress spinner animation during tests
- Verbose mode with server details
- Color-coded output behavior

**Key Code Patterns**:
```javascript
// State machine for test phases
let state = 'ping';

// Progress rendering every 50ms
setInterval(render, 50);

// Spinner based on current state
const getSpinnerFromState = inputState => 
    inputState === state ? spinner.frame() : '  ';
```

### 4.2 Protocol Reference: showwin/speedtest-go

The [showwin/speedtest-go](https://github.com/showwin/speedtest-go) is a pure Go implementation that provides valuable protocol details.

**What to Study**:
- XML parsing for server list
- HTTP client configuration
- Download/upload test mechanics
- Rate calculation (EWMA)
- Progress reporting patterns

### 4.3 Testing Patterns

This project uses table-driven tests for comprehensive coverage. The testing strategy follows Go best practices:

**Table-Driven Test Pattern:**

```go
func TestCalculateDistance(t *testing.T) {
    tests := []struct {
        name     string
        lat1, lon1 float64
        lat2, lon2 float64
        expected  float64
        delta     float64
    }{
        {"NYC to LA", 40.7128, -74.0060, 34.0522, -118.2437, 3935.7, 10.0},
        {"Same point", 40.7128, -74.0060, 40.7128, -74.0060, 0.0, 0.001},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalculateDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
            if math.Abs(got-tt.expected) > tt.delta {
                t.Errorf("CalculateDistance() = %v, want %v (±%v)", got, tt.expected, tt.delta)
            }
        })
    }
}
```

**Mock HTTP Servers:**

Use `net/http/httptest` for testing network code:

```go
func TestSelectBestServerByPing(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer ts.Close()

    server := &types.Server{URL: ts.URL}
    result, err := SelectBestServerByPing(context.Background(), []*types.Server{server}, 1)
    // Test assertions...
}
```

**Running Tests:**

```bash
# Run all tests
make test

# Generate coverage report
make test-coverage

# View coverage in browser
open coverage.html
```

**Coverage Requirements:**
- Minimum 80% code coverage (RFC Section 11.1)
- All critical paths must be tested
- Error conditions must have test coverage

### 4.3 Documentation Standards

**Go Documentation Comments**:
```go
// FetchServerList retrieves the list of available speed test servers
// from the speedtest.net API.
func FetchServerList(ctx context.Context) ([]*Server, error) {
    // Implementation
}
```

**README Structure**:
- Installation instructions
- Usage examples
- Flag documentation
- Building instructions
- Contributing guidelines

## 5. Getting Started

### 5.1 Development Environment

**Required Tools**:
- Go 1.21 or later
- golangci-lint (for code quality)
- Git (for version control)

**IDE Setup**:
- VS Code with Go extension
- GoLand (JetBrains)
- Neovim with nvim-lspconfig

### 5.2 Building and Testing

**Build:**
```bash
make build
```

**Test:**
```bash
make test
```

**Test with Coverage Report:**
```bash
make test-coverage
# Opens coverage.html with detailed coverage information
```

**Lint:**
```bash
make lint
```

**Test**:
```bash
make test
```

**Lint**:
```bash
make lint
```

### 5.3 Running the CLI

**Basic Usage**:
```bash
./speed-test
```

**With Flags**:
```bash
./speed-test --json
./speed-test --verbose
./speed-test --bytes
./speed-test --server 1234
```

## 6. Key Concepts Glossary

### Bandwidth vs Throughput

- **Bandwidth**: Maximum data transfer rate (physical limit)
- **Throughput**: Actual data transfer rate achieved

For this project, we measure throughput, which is typically lower than bandwidth.

### Mbps vs MB/s

- **Mbps** (Megabits per second): Used for network speeds (1 Mbps = 1,000,000 bits/second)
- **MB/s** (Megabytes per second): Used for file transfers (1 MB/s = 8 Mbps)

Conversion:
```
Mbps = (bytes_per_second * 8) / 1,000,000
MB/s = bytes_per_second / 1,048,576
```

### Latency vs Jitter

- **Latency**: Time for a single packet to travel (measured in ms)
- **Jitter**: Variation in latency between packets (standard deviation)

Lower is better for both, but they measure different things.

### EWMA (Exponential Weighted Moving Average)

A smoothing algorithm that gives more weight to recent observations:

```
smoothed_value = alpha * new_value + (1 - alpha) * previous_value
```

Alpha typically 0.1 (10% weight for new values).

### Haversine Formula

Calculates the great-circle distance between two points on a sphere:

```go
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
    const earthRadius = 6371.0 // km
    
    lat1Rad := lat1 * math.Pi / 180
    lat2Rad := lat2 * math.Pi / 180
    deltaLat := (lat2 - lat1) * math.Pi / 180
    deltaLon := (lon2 - lon1) * math.Pi / 180
    
    a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
        math.Cos(lat1Rad)*math.Cos(lat2Rad)*
            math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
    
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    
    return earthRadius * c
}
```

## 7. Common Patterns

### Progress Reporting with Channels

```go
// Create channel for progress updates
progress := make(chan ProgressInfo, 10)

// Start worker that sends updates
go func() {
    for i := 0; i < 100; i++ {
        progress <- ProgressInfo{Progress: float64(i) / 100}
    }
    close(progress)
}()

// Receive updates
for p := range progress {
    fmt.Printf("\rProgress: %.1f%%", p.Progress*100)
}
```

### Concurrent with Rate Limiting

```go
// Semaphore for rate limiting (max 3 concurrent)
sem := make(chan struct{}, 3)

for i := 0; i < 10; i++ {
    sem <- struct{}{} // Acquire
    go func(id int) {
        defer func() { <-sem }() // Release
        // Do work
    }(i)
}
```

### Context with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// All operations respect context cancellation
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
```

---

**Document Version**: 1.0  
**Last Updated**: 2026-01-26  
**Related Documents**:
- [RFC](rfc.md) - Technical design document
- [Task Tracking](task.md) - Implementation task list
