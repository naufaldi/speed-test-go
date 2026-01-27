# RFC: Go Network Speed Test CLI

## 1. Summary

Create a high-performance CLI tool written in Go that tests internet connection speed (download and upload) and ping latency. This tool is a rewrite of the popular Node.js [speed-test](https://github.com/sindresorhus/speed-test) CLI tool, implementing the speedtest.net protocol from scratch in pure Go, leveraging Go's superior performance characteristics for network operations and creating a standalone binary distribution.

## 2. Motivation

The original Node.js speed-test tool has gained significant popularity (3.9k stars) but requires Node.js runtime and wraps the Ookla Speedtest CLI binary, adding unnecessary dependency overhead for a simple CLI utility. A Go rewrite offers:

- **Single Binary Distribution**: No runtime dependencies or binary downloads, easy installation via `go install` or direct binary downloads
- **Pure Go Implementation**: Custom implementation of speedtest.net protocol without external wrappers
- **Superior Performance**: Go's efficient concurrency model and low-level network handling
- **Cross-Platform Compilation**: Native support for all major platforms with zero modification
- **Familiar CLI Experience**: POSIX-compliant flags and intuitive output matching the original tool

## 3. Goals

### 3.1 Primary Goals
- Implement speedtest.net protocol from scratch in pure Go
- Implement ping latency measurement to speed test servers
- Implement download speed test with real-time progress updates
- Implement upload speed test with real-time progress updates
- Provide human-readable output with optional JSON mode
- Provide unit conversion (Mbps vs MB/s)
- Auto-select nearest/fastest server or allow manual server selection

### 3.2 Secondary Goals
- Support for custom server configurations
- Verbose mode with server information (location, distance, sponsor)
- Unit tests with high coverage (80%+)
- CI/CD pipeline for automated builds and releases
- Progress indicators with real-time speed updates

## 4. Non-Goals

- Building a GUI or web interface (CLI-only tool)
- Implementing speedtest.net account integration
- Historical data storage or result tracking over time
- Supporting speedtest.net result submission to their service
- Implementing complex network diagnostics beyond speed and ping
- Wrapping the Ookla CLI binary (pure Go implementation)

## 5. Implementation Approach

### 5.1 Custom Protocol Implementation

We will **implement the speedtest.net protocol from scratch** in pure Go, not using external speedtest libraries as runtime dependencies.

#### Reference Materials:

1. **Primary CLI/UX Reference**: [sindresorhus/speed-test](https://github.com/sindresorhus/speed-test/blob/main/cli.js)
   - **THIS IS THE TOOL WE ARE REWRITING** - A Node.js CLI implementation of speedtest.net
   - Key features to match:
     - Flags: `--json` (`-j`), `--bytes` (`-b`), `--verbose` (`-v`)
     - Output format: Ping, Download, Upload with real-time progress
     - Spinner animation during active test phases
     - Verbose mode showing server host, location, country, and distance
     - JSON output mode for scripting
     - Color-coded output (cyan for results, yellow for in-progress)
     - Error handling with clear messages

2. **Secondary Protocol Reference**: [showwin/speedtest-go](https://github.com/showwin/speedtest-go)
   - Pure Go implementation to study speedtest.net protocol
   - Understanding server discovery XML format
   - Download/upload test mechanics
   - Bandwidth calculation and rate smoothing

3. **Output Format Reference**: [speedtest-net](https://github.com/ddsol/speedtest.net)
   - Ookla CLI JSON output format for JSON schema compatibility
   - Progress event types and data structures

#### Protocol Implementation:
- Server discovery via speedtest.net XML API
- User location detection
- Ping measurement (HTTP/TCP/ICMP)
- Multi-threaded download testing
- Multi-threaded upload testing
- Bandwidth calculation and rate smoothing

### 5.2 Dependencies Strategy

**Runtime Dependencies**:
- [Cobra](https://github.com/spf13/cobra) - CLI framework (Apache 2.0)
- Go standard library only (`net/http`, `net`, `context`, `time`, `encoding/xml`, `encoding/json`)

**Development/Testing Dependencies**:
- [Testify](https://github.com/stretchr/testify) - Testing assertions (MIT)
- [GoMock](https://github.com/golang/mock) - Code-generated mocks (Apache 2.0)

**Reference Only** (not imported):
- **sindresorhus/speed-test CLI** (`/Users/mac/WebApps/oss/speedtest.net`) - Primary reference for CLI design and UX patterns
- showwin/speedtest-go - Protocol implementation details

## 6. Technical Design

### 6.1 Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    CLI Entry Point                          │
│                        main.go                              │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                     Command Handler                         │
│                   cmd/root.go (Cobra)                       │
└─────────────────────┬───────────────────────────────────────┘
                      │
          ┌───────────┴───────────┐
          ▼                       ▼
┌─────────────────┐     ┌─────────────────────────┐
│   Test Runner   │     │    Output Formatter     │
│  internal/test  │     │   internal/output       │
│                 │     │                         │
│ - Ping test     │     │ - Human readable        │
│ - Download test │     │ - JSON output           │
│ - Upload test   │     │ - Verbose server info   │
└────────┬────────┘     └─────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│                  Network Layer                              │
│              internal/network/client.go                     │
└─────────────────────┬───────────────────────────────────────┘
                      │
          ┌───────────┴───────────┐
          ▼                       ▼
┌─────────────────┐     ┌─────────────────────────┐
│ Server Selector │     │ Data Transfer Engine    │
│ internal/server │     │ internal/transfer       │
│                 │     │                         │
│ - Fetch servers │     │ - Download handler      │
│ - Sort by dist  │     │ - Upload handler        │
│ - Select best   │     │ - Rate calculation      │
└─────────────────┘     └─────────────────────────┘
```

### 6.2 Key Components

#### 6.2.1 CLI Framework (Cobra)

```go
// cmd/root.go
var rootCmd = &cobra.Command{
    Use:   "speed-test",
    Short: "Test your internet connection speed and ping",
    Long: `Test your internet connection speed and ping using speedtest.net from the CLI.
    
Supports multiple output formats and configuration options.`,
    RunE: runSpeedTest,
}

var (
    jsonFlag      bool
    bytesFlag     bool
    verboseFlag   bool
    serverIDFlag  string
    timeoutFlag   time.Duration
)

func init() {
    rootCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output the result as JSON")
    rootCmd.Flags().BoolVarP(&bytesFlag, "bytes", "b", false, "Output the result in megabytes per second (MB/s)")
    rootCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Output more detailed information")
    rootCmd.Flags().StringVarP(&serverIDFlag, "server", "s", "", "Specify a server ID to use")
    rootCmd.Flags().DurationVarP(&timeoutFlag, "timeout", "t", 30*time.Second, "Timeout for the speed test")
}
```

#### 6.2.2 Speed Test Logic

```go
// internal/test/runner.go
type SpeedTestResult struct {
    Timestamp time.Time        `json:"timestamp"`
    Ping      PingResult       `json:"ping"`
    Download  TransferResult   `json:"download"`
    Upload    TransferResult   `json:"upload"`
    Server    *ServerInfo      `json:"server,omitempty"`
    Interface *InterfaceInfo   `json:"interface,omitempty"`
    ISP       string           `json:"isp,omitempty"`
}

type PingResult struct {
    Jitter  float64 `json:"jitter"`  // milliseconds
    Latency float64 `json:"latency"` // milliseconds
}

type TransferResult struct {
    Bandwidth int64 `json:"bandwidth"` // bytes per second
    Bytes     int64 `json:"bytes"`
    Elapsed   int64 `json:"elapsed"` // milliseconds
}

type ServerInfo struct {
    ID       string  `json:"id"`
    Host     string  `json:"host"`
    Name     string  `json:"name"`
    Location string  `json:"location"`
    Country  string  `json:"country"`
    Sponsor  string  `json:"sponsor"`
    Distance float64 `json:"distance"` // in km
}

type InterfaceInfo struct {
    InternalIP string `json:"internalIp"`
    ExternalIP string `json:"externalIp"`
    Name       string `json:"name"`
    MACAddr    string `json:"macAddr"`
}

type SpeedTestRunner struct {
    client      *http.Client
    timeout     time.Duration
    serverList  []*Server
    dataManager *DataManager
}

func (r *SpeedTestRunner) Run(ctx context.Context) (*SpeedTestResult, error)
```

#### 6.2.3 Data Transfer Engine

```go
// internal/transfer/engine.go
type TransferEngine struct {
    nThreads           int
    captureTime        time.Duration
    rateCaptureFreq    time.Duration
}

func (e *TransferEngine) RunDownloadTest(ctx context.Context, serverURL string, progress chan<- TransferProgress) error
func (e *TransferEngine) RunUploadTest(ctx context.Context, serverURL string, dataSize int64, progress chan<- TransferProgress) error

type TransferProgress struct {
    Rate         float64   // current transfer rate in bytes per second
    BytesTotal   int64
    BytesCurrent int64
    Elapsed      time.Duration
}
```

### 6.3 Server Discovery and Selection

Following the speedtest.net protocol for server list retrieval:

```go
// internal/server/discovery.go
type Server struct {
    URL      string  `xml:"url,attr"`
    Lat      string  `xml:"lat,attr"`
    Lon      string  `xml:"lon,attr"`
    Name     string  `xml:"name,attr"`
    Country  string  `xml:"country,attr"`
    Sponsor  string  `xml:"sponsor,attr"`
    ID       string  `xml:"id,attr"`
    Host     string  `xml:"host,attr"`
    Distance float64 `json:"distance"`
}

type ServerList struct {
    Servers []*Server `xml:"servers>server"`
}

func FetchServerList(ctx context.Context) ([]*Server, error)
func SelectBestServer(ctx context.Context, servers []*Server) (*Server, error)
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 // Haversine formula
```

### 6.4 User Location Detection

```go
// internal/location/client.go
type UserLocation struct {
    IP        string  `xml:"ip,attr"`
    Latitude  float64 `xml:"lat,attr"`
    Longitude float64 `xml:"lon,attr"`
    ISP       string  `xml:"isp,attr"`
}

func DetectUserLocation(ctx context.Context) (*UserLocation, error)
```

### 6.5 Output Formatting

Following the sindresorhus/speed-test output format pattern:

```go
// internal/output/formatter.go
type ResultFormatter struct {
    useBytes   bool
    useJSON    bool
    useVerbose bool
    useColor   bool  // Match original's colored output
}

func (f *ResultFormatter) Format(result *SpeedTestResult) string
func (f *ResultFormatter) FormatJSON(result *SpeedTestResult) ([]byte, error)

// Output states matching original CLI behavior
type OutputState string

const (
    StateIdle     OutputState = "idle"
    StatePing     OutputState = "ping"
    StateDownload OutputState = "download"
    StateUpload   OutputState = "upload"
    StateDone     OutputState = "done"
)

func formatSpeed(bytesPerSecond float64, useBytes bool) string {
    if useBytes {
        // Convert to MB/s
        mbps := bytesPerSecond / (1024 * 1024)
        return fmt.Sprintf("%.2f MBps", mbps)
    }
    // Convert to Mbps (matches original: "Mbps")
    mbps := (bytesPerSecond * 8) / (1000 * 1000)
    return fmt.Sprintf("%.2f Mbps", mbps)
}

// Format matches sindresorhus/speed-test output layout:
//     Ping     24.5 ms
// Download  95.32 Mbps
//   Upload   23.45 Mbps
```

## 7. CLI Interface Design

### 7.1 Usage Examples

```bash
# Basic speed test
$ speed-test

# JSON output for scripting
$ speed-test --json

# Output in MB/s instead of Mbps
$ speed-test --bytes

# Verbose output with server details
$ speed-test --verbose

# Specify a server by ID
$ speed-test --server 1234

# Set timeout
$ speed-test --timeout 60s

# Combined flags
$ speed-test --json --verbose

# Help
$ speed-test --help
```

### 7.2 Output Examples

#### Human Readable Output (Default)
```
 Ping     24.5 ms
 Download 95.32 Mbps
 Upload   23.45 Mbps
```

#### Verbose Output
```
 Ping     24.5 ms
 Download 95.32 Mbps
 Upload   23.45 Mbps

 Server   speedtest.server.com
 Location New York, NY
 Country  United States
 Sponsor  Example ISP
 Distance 1,245.3 km
```

#### JSON Output
```json
{
  "timestamp": "2026-01-26T10:30:00Z",
  "ping": {
    "jitter": 1.022,
    "latency": 24.5
  },
  "download": {
    "bandwidth": 11915965,
    "bytes": 128794451,
    "elapsed": 10804
  },
  "upload": {
    "bandwidth": 2931294,
    "bytes": 28447808,
    "elapsed": 9703
  },
  "server": {
    "id": "1234",
    "host": "speedtest.server.com:8080",
    "name": "New York Server",
    "location": "New York, NY",
    "country": "United States",
    "sponsor": "Example ISP",
    "distance": 1245.3
  },
  "interface": {
    "internalIp": "192.168.1.100",
    "externalIp": "203.0.113.42",
    "name": "en0",
    "macAddr": "00:1B:63:84:45:E6"
  },
  "isp": "Example Internet Provider"
}
```

### 7.4 Matching sindresorhus/speed-test UX

This Go implementation aims to **exactly match** the original sindresorhus/speed-test CLI behavior:

**Progress Display:**
- Spinner animation during active test phases (ping, download, upload)
- Real-time speed updates (matches original's 50ms render interval)
- Color-coded output:
  - Cyan: Completed/measured values
  - Yellow: In-progress values
  - Dim gray: Unit labels and spinner

**Output Layout:**
- Three-line format with aligned columns
- Spinner prefix showing active test phase
- Unit labels in dim gray text (e.g., "ms", "Mbps", "MBps")

**State Machine:**
```
idle → ping → download → upload → done
```
- Each phase has distinct visual indicator
- Progress updates flow through all phases

**Error Handling:**
- Network errors show friendly messages
- JSON mode returns structured error responses

### 7.3 Flag Reference

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | `-j` | boolean | false | Output the result as JSON |
| `--bytes` | `-b` | boolean | false | Output in megabytes per second (MBps) instead of megabits per second (Mbps) |
| `--verbose` | `-v` | boolean | false | Output more detailed information including server details |
| `--server` | `-s` | string | "" | Specify a server ID to use for testing |
| `--timeout` | `-t` | duration | 30s | Timeout for the entire speed test |
| `--help` | `-h` | boolean | false | Show help information |
| `--version` | `-V` | boolean | false | Show version information |

## 8. Speedtest.net Protocol Implementation

### 8.1 Protocol Endpoints

**Server List Discovery**:
```
GET https://www.speedtest.net/api/js/servers?engine=js&limit=10
```

**User Location**:
```
GET http://speedtest.net/speedtest-config.php
```

**Ping Test**:
```
GET http://server-host/speedtest/latency.txt
```

**Download Test**:
```
GET http://server-host/speedtest/random{size}x{size}.jpg
```

**Upload Test**:
```
POST http://server-host/speedtest/upload.php
Content-Type: application/octet-stream
```

### 8.2 Test Phases

1. **Initialization Phase**
   - Fetch user location
   - Fetch server list
   - Calculate distances
   - Sort servers by distance

2. **Server Selection Phase**
   - Ping top N closest servers
   - Select server with lowest latency

3. **Ping Test Phase**
   - Send multiple ping requests (5-10)
   - Calculate average latency and jitter

4. **Download Test Phase**
   - Start with 1 thread
   - Scale up to optimal thread count
   - Download test data for ~15 seconds
   - Calculate bandwidth

5. **Upload Test Phase**
   - Start with 1 thread
   - Scale up to optimal thread count
   - Upload test data for ~10 seconds
   - Calculate bandwidth

## 9. Dependencies

### 9.1 Runtime Dependencies

| Dependency | Purpose | License |
|------------|---------|---------|
| [Cobra](https://github.com/spf13/cobra) | CLI framework | Apache 2.0 |
| Go standard library | All network operations | BSD-style |

### 9.2 Development Dependencies

| Dependency | Purpose | License |
|------------|---------|---------|
| [Testify](https://github.com/stretchr/testify) | Testing assertions | MIT |
| [GoMock](https://github.com/golang/mock) | Code-generated mocks | Apache 2.0 |

### 9.3 Reference Materials (Not Imported)

| Resource | Purpose |
|----------|---------|
| [showwin/speedtest-go](https://github.com/showwin/speedtest-go) | Protocol implementation reference |
| [speedtest-net](https://github.com/ddsol/speedtest.net) | Output format reference |
| [sindresorhus/speed-test](https://github.com/sindresorhus/speed-test) | UX inspiration |

## 10. Implementation Plan

### Phase 1: Project Setup (Week 1)
- [ ] Initialize Go module: `go mod init github.com/user/speed-test-go`
- [ ] Set up project structure (cmd/, internal/, pkg/)
- [ ] Configure Cobra CLI framework
- [ ] Set up GitHub Actions CI/CD
- [ ] Configure linters (golangci-lint, gofumpt)
- [ ] Create Makefile with common targets (build, test, lint)

### Phase 2: Core Network Layer (Week 2)
- [ ] Implement server discovery from speedtest.net API
- [ ] Implement user location detection
- [ ] Implement distance calculation (Haversine formula)
- [ ] Create HTTP client with proper timeouts
- [ ] Implement ping measurement (HTTP-based)
- [ ] Write unit tests for network layer

### Phase 3: Speed Test Implementation (Week 3)
- [ ] Implement download speed test with chunked transfer
- [ ] Implement upload speed test with POST requests
- [ ] Add concurrent transfer support (multi-threaded)
- [ ] Implement rate calculation (EWMA for smoothing)
- [ ] Add progress reporting via channels
- [ ] Write unit tests for transfer engine

### Phase 4: CLI and Output (Week 4)
- [ ] Complete CLI flag implementation
- [ ] Implement human-readable output formatter
- [ ] Implement JSON output formatter
- [ ] Add verbose mode with server information
- [ ] Add progress indicators (spinner/bar)
- [ ] Write integration tests for CLI

### Phase 5: Polish and Release (Week 5)
- [ ] Add version command with build info
- [ ] Write comprehensive README
- [ ] Create installation documentation
- [ ] Cross-platform testing (Linux, macOS, Windows)
- [ ] Set up goreleaser for automated releases
- [ ] Final code review and optimization

## 11. Testing Strategy

### 11.1 Unit Testing
- Minimum 80% code coverage
- Table-driven tests for core logic
- Mock external dependencies (HTTP, time)
- Test edge cases (timeouts, network errors, malformed responses)

### 11.2 Integration Testing
- Full CLI command execution tests
- JSON output schema validation
- Cross-platform behavior verification
- Performance benchmarking

### 11.3 Manual Testing
- Test with various network conditions
- Verify server selection accuracy
- Test all flag combinations
- Verify output formatting on different terminals

## 12. Performance Considerations

### 12.1 Concurrency Model
- Use goroutines for concurrent download/upload streams
- Implement worker pool pattern for multi-threaded transfers
- Context-based cancellation for responsive interrupts
- Channel-based progress reporting

### 12.2 Memory Management
- Stream large transfers without loading entire files
- Use buffered I/O with appropriate buffer sizes (32KB-64KB)
- Avoid unnecessary allocations in hot paths
- Use sync.Pool for frequently allocated objects

### 12.3 Network Optimization
- Configure appropriate HTTP timeouts (connect: 5s, read: 10s, write: 10s)
- Enable HTTP/2 when server supports it
- Use connection keep-alive for multiple requests
- Implement exponential backoff for retries

## 13. Security Considerations

### 13.1 Network Security
- Validate all server certificates for HTTPS
- Use TLS 1.2+ only (no weak protocols)
- Sanitize all user input and server responses
- Prevent path traversal attacks

### 13.2 Data Safety
- Don't store sensitive data (IP addresses, location) to disk
- Use constant-time comparisons for sensitive operations
- Implement proper error handling without information leakage
- No telemetry or data collection

### 13.3 Supply Chain
- Pin all dependency versions in go.mod
- Use Go's module verification (sumdb)
- Regular dependency vulnerability scanning (govulncheck)
- Minimal dependency surface area

## 14. Distribution

### 14.1 Installation Methods

1. **Go Install**:
   ```bash
   go install github.com/user/speed-test-go@latest
   ```

2. **Homebrew** (macOS/Linux):
   ```bash
   brew install speed-test-go
   ```

3. **Pre-built Binaries**:
   - GitHub Releases with binaries for:
     - Linux (amd64, arm64, arm)
     - macOS (amd64, arm64)
     - Windows (amd64)
     - FreeBSD (amd64)

4. **Docker**:
   ```bash
   docker run --rm ghcr.io/user/speed-test-go
   ```

### 14.2 Release Process
- Semantic versioning (SemVer): v1.0.0, v1.1.0, etc.
- Automated builds via GitHub Actions
- Checksum verification (SHA256) for all releases
- Optional: Signed releases using cosign

## 15. Future Enhancements

### 15.1 Post-MVP Features
- [ ] Configuration file support (~/.speed-test-go.yaml)
- [ ] Historical result tracking (optional local storage)
- [ ] Export to various formats (CSV, HTML report)
- [ ] Docker container with minimal image size
- [ ] Support for alternative test servers
- [ ] Packet loss measurement
- [ ] ICMP ping support (requires privileges)
- [ ] Bandwidth throttling for testing

### 15.2 Advanced Features
- [ ] Web dashboard mode (embedded HTTP server)
- [ ] Continuous monitoring mode
- [ ] Comparison with previous tests
- [ ] Network quality score calculation
- [ ] Integration with monitoring systems (Prometheus)

## 16. Project Structure

```
speed-test-go/
├── cmd/
│   └── root.go              # CLI entry point and command definitions
├── internal/
│   ├── location/
│   │   └── client.go        # User location detection
│   ├── network/
│   │   └── client.go        # HTTP client configuration
│   ├── output/
│   │   ├── formatter.go     # Output formatting (human/JSON)
│   │   └── progress.go      # Progress reporting
│   ├── server/
│   │   ├── discovery.go     # Server list fetching
│   │   ├── selection.go     # Server selection logic
│   │   └── distance.go      # Distance calculation
│   ├── test/
│   │   ├── runner.go        # Main test runner
│   │   ├── ping.go          # Ping test implementation
│   │   └── result.go        # Result types
│   └── transfer/
│       ├── download.go      # Download test engine
│       ├── upload.go        # Upload test engine
│       └── rate.go          # Rate calculation (EWMA)
├── pkg/
│   └── types/               # Public type definitions
├── .github/
│   └── workflows/
│       ├── ci.yml           # Continuous integration
│       ├── release.yml      # Automated releases
│       └── codeql.yml       # Security scanning
├── docs/
│   └── rfc.md               # This RFC document
├── .golangci.yml            # Linter configuration
├── go.mod
├── go.sum
├── main.go                  # Application entry point
├── Makefile                 # Build automation
├── README.md                # User documentation
└── LICENSE                  # MIT License
```

## 17. References

### 17.1 Primary Reference (The Tool We Are Rewriting)

**sindresorhus/speed-test** - Node.js CLI Implementation
- GitHub: https://github.com/sindresorhus/speed-test
- CLI Code: https://github.com/sindresorhus/speed-test/blob/main/cli.js
- **This is the EXACT tool we are rewriting from Node.js to Go**

This implementation provides the reference for:
- CLI flag interface (`--json`, `--bytes`, `--verbose`)
- Output format and layout (Ping, Download, Upload)
- Progress spinner animation during tests
- Verbose mode with server details (host, location, country, distance)
- Color-coded output (cyan for completed, yellow for in-progress)
- Error handling patterns

### 17.2 Secondary Reference (Protocol Implementation)

**showwin/speedtest-go** - Pure Go Protocol Implementation
- GitHub: https://github.com/showwin/speedtest-go
- Study source code to understand speedtest.net protocol details
- Server discovery, download/upload mechanics, rate calculation

### 17.3 Additional References

- **speedtest-net** - Ookla CLI Wrapper (JSON format reference)
  - GitHub: https://github.com/ddsol/speedtest.net

- **Cobra** - Go CLI Framework
  - GitHub: https://github.com/spf13/cobra

- **Speedtest.net** - Official Service
  - https://www.speedtest.net

## 18. Glossary

- **Mbps**: Megabits per second (standard unit for network speed), 1 Mbps = 125,000 bytes/s
- **MB/s**: Megabytes per second, 1 MB/s = 8 Mbps
- **Ping**: Round-trip time latency measurement in milliseconds
- **Jitter**: Variation in packet delay (standard deviation of latency)
- **EWMA**: Exponential Weighted Moving Average for rate smoothing
- **Haversine**: Formula for calculating great-circle distance between two points on Earth
- **Bandwidth**: Data transfer rate in bytes per second

## 19. Success Criteria

The project will be considered successful when:

1. **Functionality**: All core features work correctly
   - Accurate ping measurement (±5ms of Ookla CLI)
   - Accurate download speed (±10% of Ookla CLI)
   - Accurate upload speed (±10% of Ookla CLI)
   - JSON output matches expected schema

2. **Performance**: Fast and efficient execution
   - Total test time < 30 seconds
   - Memory usage < 50MB
   - Binary size < 10MB

3. **Quality**: High code quality standards
   - 80%+ code coverage
   - Zero critical security vulnerabilities
   - All linters passing

4. **Usability**: Easy to install and use
   - Single command installation
   - Clear, readable output
   - Comprehensive documentation

5. **Compatibility**: Works across platforms
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)

## 20. Open Questions

1. **Thread Count**: What's the optimal number of concurrent threads for download/upload? (To be determined through testing)
2. **Server Count**: How many servers should we ping initially? (Tentative: 5 closest)
3. **Test Duration**: How long should each test phase run? (Tentative: 15s download, 10s upload)
4. **Progress Updates**: How frequently should we update progress? (Tentative: 100ms)
5. **Retry Logic**: How many retries on network errors? (Tentative: 3 retries with exponential backoff)

These questions will be answered during implementation and testing phases.

---

**RFC Status**: Draft  
**Author**: Generated via AI Assistant  
**Date**: 2026-01-26  
**Version**: 1.0
