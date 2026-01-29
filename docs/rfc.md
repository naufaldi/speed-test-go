# Technical Documentation: Go Network Speed Test CLI

> A pure Go implementation of the speedtest.net protocol for testing internet connection speed and latency.

## Table of Contents

1. [Overview](#1-overview)
2. [Architecture](#2-architecture)
3. [Test Flow](#3-test-flow)
4. [Core Components](#4-core-components)
5. [Data Types](#5-data-types)
6. [CLI Interface](#6-cli-interface)
7. [Technical Details](#7-technical-details)
8. [Building & Testing](#8-building--testing)

---

## 1. Overview

Speed Test CLI is a high-performance command-line tool that measures:
- **Ping latency** - Round-trip time to speed test servers
- **Download speed** - Bandwidth for receiving data
- **Upload speed** - Bandwidth for sending data

### Key Features

| Feature | Description |
|---------|-------------|
| Pure Go | No external binary dependencies or runtime requirements |
| Multi-threaded | 4 concurrent threads for download, 2 for upload |
| Smart Server Selection | Ping-based selection from closest servers |
| Multiple Output Formats | Human-readable, JSON, and verbose modes |
| Cross-Platform | Linux, macOS, Windows binaries |

---

## 2. Architecture

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Entry                               │
│                        main.go                                  │
│                           │                                     │
│                           ▼                                     │
│                    cmd/root.go                                  │
│                    (Cobra CLI)                                  │
└───────────────────────────┬─────────────────────────────────────┘
                            │
            ┌───────────────┼───────────────┐
            ▼               ▼               ▼
┌───────────────────┐ ┌───────────────┐ ┌───────────────────┐
│   Test Runner     │ │    Output     │ │     Network       │
│  internal/test    │ │internal/output│ │ internal/network  │
│                   │ │               │ │                   │
│  - Orchestrates   │ │  - Formatter  │ │  - HTTP Client    │
│    all tests      │ │  - Progress   │ │    configuration  │
└─────────┬─────────┘ └───────────────┘ └───────────────────┘
          │
          ├──────────────┬──────────────┬──────────────┐
          ▼              ▼              ▼              ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│   Location   │ │    Server    │ │   Transfer   │ │     Test     │
│   Detection  │ │  Discovery   │ │    Engine    │ │     Ping     │
│              │ │              │ │              │ │              │
│internal/     │ │internal/     │ │internal/     │ │internal/     │
│location      │ │server        │ │transfer      │ │test          │
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
```

### 2.2 Package Structure

```
speed-test-go/
├── main.go                 # Entry point → calls cmd.Execute()
├── cmd/
│   ├── root.go             # Main CLI command (Cobra)
│   └── version.go          # Version subcommand
├── internal/               # Private packages
│   ├── location/           # User IP/location detection
│   │   └── client.go       # Fetches speedtest-config.php
│   ├── network/            # HTTP client factory
│   │   └── client.go       # Configured HTTP clients
│   ├── output/             # Output formatting
│   │   ├── formatter.go    # Human/JSON output
│   │   └── progress.go     # Progress indicators
│   ├── server/             # Server management
│   │   ├── discovery.go    # Fetch server list (JSON API)
│   │   ├── selection.go    # Ping-based server selection
│   │   └── distance.go     # Haversine distance calculation
│   ├── test/               # Test execution
│   │   ├── runner.go       # Main test orchestrator
│   │   └── ping.go         # Ping/latency measurement
│   └── transfer/           # Bandwidth testing
│       ├── download.go     # Multi-threaded download test
│       ├── upload.go       # Multi-threaded upload test
│       ├── rate.go         # EWMA rate smoothing
│       └── types.go        # Transfer types
└── pkg/
    └── types/              # Public type definitions
        └── types.go        # SpeedTestResult, Server, etc.
```

---

## 3. Test Flow

### 3.1 Execution Sequence

```
┌──────────────────────────────────────────────────────────────────┐
│                     Speed Test Execution Flow                     │
└──────────────────────────────────────────────────────────────────┘

     ┌─────────────┐
     │   START     │
     └──────┬──────┘
            │
            ▼
┌─────────────────────┐    GET speedtest-config.php
│  1. Detect User     │◄──────────────────────────────┐
│     Location        │    Returns: IP, lat/lon, ISP  │
└──────────┬──────────┘                               │
           │                                          │
           ▼                              ┌───────────┴───────────┐
┌─────────────────────┐                   │  speedtest.net API    │
│  2. Fetch Server    │ GET /api/js/servers?engine=js&limit=10    │
│     List            │◄──────────────────┴───────────────────────┘
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  3. Calculate       │    Haversine formula
│     Distances       │    user coords → server coords
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  4. Select Best     │    Ping top N closest servers
│     Server          │    Choose lowest latency
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  5. Run Ping Test   │    5 HTTP requests to /speedtest/latency.txt
│                     │    Calculate avg latency + jitter
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  6. Run Download    │    4 threads × 15 seconds
│     Test            │    Download random images
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  7. Run Upload      │    2 threads × 20 seconds
│     Test            │    POST random data
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  8. Format & Output │    Human-readable or JSON
│     Results         │
└──────────┬──────────┘
           │
           ▼
     ┌─────────────┐
     │    END      │
     └─────────────┘
```

### 3.2 Server Selection Algorithm

```
                    ┌─────────────────────┐
                    │ Fetch 10 servers    │
                    │ from speedtest.net  │
                    └──────────┬──────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │ Calculate distance  │
                    │ to each server      │
                    │ (Haversine formula) │
                    └──────────┬──────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │ Sort by distance    │
                    │ (closest first)     │
                    └──────────┬──────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │ Take top N servers  │
                    │ (default: 5)        │
                    └──────────┬──────────┘
                               │
           ┌───────────────────┼───────────────────┐
           ▼                   ▼                   ▼
    ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
    │  Ping       │     │  Ping       │     │  Ping       │
    │  Server 1   │     │  Server 2   │     │  Server N   │
    │  (3 pings)  │     │  (3 pings)  │     │  (3 pings)  │
    └──────┬──────┘     └──────┬──────┘     └──────┬──────┘
           │                   │                   │
           └───────────────────┼───────────────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │ Select server with  │
                    │ lowest avg latency  │
                    └─────────────────────┘
```

---

## 4. Core Components

### 4.1 Test Runner (`internal/test/runner.go`)

The central orchestrator that coordinates all test phases:

```go
type Runner struct {
    maxServers       int    // Max servers to consider
    serverID         string // Specific server ID (optional)
    numServersToTest int    // Servers to ping for selection
}

func (r *Runner) Run(ctx context.Context) (*types.SpeedTestResult, error)
```

**Responsibilities:**
1. Detect user location via speedtest.net config API
2. Fetch and sort server list by distance
3. Select best server (by ping or explicit ID)
4. Execute ping, download, and upload tests
5. Aggregate results into `SpeedTestResult`

### 4.2 Server Discovery (`internal/server/discovery.go`)

Fetches available speed test servers from the speedtest.net JSON API:

```go
const serverListURL = "https://www.speedtest.net/api/js/servers?engine=js&limit=10"

func FetchServerList(ctx context.Context) ([]*types.Server, error)
```

### 4.3 Server Selection (`internal/server/selection.go`)

Implements intelligent server selection:

```go
func SelectBestServerByPing(ctx context.Context, servers []*types.Server, numServers int) (*types.Server, error)
```

**Algorithm:**
1. Take top N closest servers
2. Ping each server 3 times concurrently (max 3 concurrent pings)
3. Calculate average latency per server
4. Return server with lowest latency

### 4.4 Transfer Engine (`internal/transfer/`)

#### Download Test
- **Threads:** 4 concurrent goroutines
- **Duration:** 15 seconds
- **URLs:** Speedtest.net random images, with fallbacks to CDNs (Cloudflare, Google)
- **Buffer Size:** 32KB read buffer

#### Upload Test
- **Threads:** 2 concurrent goroutines  
- **Duration:** 20 seconds
- **Payload:** 1MB random data per request
- **Endpoints:** Speedtest.net upload.php, with fallbacks

### 4.5 Rate Calculator (`internal/transfer/rate.go`)

Uses **Exponential Weighted Moving Average (EWMA)** for rate smoothing:

```go
type EWMA struct {
    alpha float64  // Weight for new values (0.1 = 10%)
    value float64  // Current smoothed value
}
```

**Formula:** `smoothed = α × new_value + (1-α) × old_value`

---

## 5. Data Types

### 5.1 Main Result Type

```go
// pkg/types/types.go
type SpeedTestResult struct {
    Timestamp time.Time      `json:"timestamp"`
    Ping      PingResult     `json:"ping"`
    Download  TransferResult `json:"download"`
    Upload    TransferResult `json:"upload"`
    Server    *ServerInfo    `json:"server,omitempty"`
    Interface *InterfaceInfo `json:"interface,omitempty"`
    ISP       string         `json:"isp,omitempty"`
}
```

### 5.2 Component Types

```go
type PingResult struct {
    Jitter  float64 `json:"jitter"`   // ms - variation in latency
    Latency float64 `json:"latency"`  // ms - average round-trip time
}

type TransferResult struct {
    Bandwidth int64 `json:"bandwidth"` // bytes per second
    Bytes     int64 `json:"bytes"`     // total bytes transferred
    Elapsed   int64 `json:"elapsed"`   // milliseconds
}

type ServerInfo struct {
    ID       string  `json:"id"`
    Host     string  `json:"host"`
    Name     string  `json:"name"`
    Country  string  `json:"country"`
    Sponsor  string  `json:"sponsor"`
    Distance float64 `json:"distance"` // km
}
```

---

## 6. CLI Interface

### 6.1 Command Structure

Built with [Cobra](https://github.com/spf13/cobra):

```go
var rootCmd = &cobra.Command{
    Use:   "speed-test",
    Short: "Test your internet connection speed and ping",
    RunE:  runSpeedTest,
}
```

### 6.2 Available Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | `-j` | bool | false | Output as JSON |
| `--bytes` | `-b` | bool | false | Show speed in MB/s instead of Mbps |
| `--verbose` | `-v` | bool | false | Include server details |
| `--server` | `-s` | string | "" | Use specific server ID |
| `--servers` | `-n` | int | 5 | Number of servers to ping |
| `--timeout` | `-t` | duration | 30s | Overall timeout |
| `--progress` | `-p` | bool | false | Show progress during test |

### 6.3 Output Formats

**Human-readable (default):**
```
      Ping 24.5 ms
  Download 95.32 Mbps
    Upload 23.45 Mbps
```

**Verbose mode (`-v`):**
```
      Ping 24.5 ms
  Download 95.32 Mbps
    Upload 23.45 Mbps

    Server   speedtest.server.com
  Location   New York (United States)
  Distance   1245.3 km
```

**JSON mode (`-j`):**
```json
{
  "timestamp": "2026-01-28T10:30:00Z",
  "ping": { "jitter": 1.022, "latency": 24.5 },
  "download": { "bandwidth": 11915965, "bytes": 128794451, "elapsed": 10804 },
  "upload": { "bandwidth": 2931294, "bytes": 28447808, "elapsed": 9703 }
}
```

---

## 7. Technical Details

### 7.1 Distance Calculation

Uses the **Haversine formula** to calculate great-circle distance:

```go
// internal/server/distance.go
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
    // Earth radius in km
    const R = 6371.0
    
    dLat := (lat2 - lat1) * math.Pi / 180
    dLon := (lon2 - lon1) * math.Pi / 180
    
    a := math.Sin(dLat/2)*math.Sin(dLat/2) +
         math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
         math.Sin(dLon/2)*math.Sin(dLon/2)
    
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    return R * c
}
```

### 7.2 Concurrency Model

```
┌────────────────────────────────────────────────────────────────┐
│                      Download Test                              │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│   Thread 1 ───────────────────────────────────────────────►    │
│   Thread 2 ───────────────────────────────────────────────►    │
│   Thread 3 ───────────────────────────────────────────────►    │
│   Thread 4 ───────────────────────────────────────────────►    │
│                                                                │
│   ◄─────────────── 15 seconds ────────────────►               │
│                                                                │
│   Shared: RateCalculator (mutex protected)                     │
│   Communication: Progress channel                              │
│   Cancellation: Context with timeout                           │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### 7.3 HTTP Client Configuration

```go
// internal/network/client.go
func DownloadClient() *http.Client {
    return &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    }
}
```

### 7.4 Fallback URLs

When speedtest.net servers don't respond, the tool falls back to:

**Download:**
- `https://speed.cloudflare.com/__down?bytes=1000000`
- `https://dl.google.com/dl/testbed/testfile1000k.bin`

**Upload:**
- `https://postman-echo.com/post`
- `https://httpbin.org/post`

---

## 8. Building & Testing

### 8.1 Commands

```bash
# Build binary
make build

# Run all tests with coverage
make test

# Run single test
go test -v -run TestName ./path/to/package

# Run linter
make lint

# Generate coverage report
make test-coverage
```

### 8.2 Dependencies

**Runtime:**
- `github.com/spf13/cobra` - CLI framework

**Dev/Test:**
- Go standard library only (no test framework dependencies)

### 8.3 Performance Targets

| Metric | Target |
|--------|--------|
| Binary Size | < 15 MB |
| Memory Usage | < 50 MB |
| Test Duration | < 30 seconds |
| Code Coverage | > 80% |

---

## References

- [sindresorhus/speed-test](https://github.com/sindresorhus/speed-test) - Original Node.js CLI (UI/UX reference)
- [showwin/speedtest-go](https://github.com/showwin/speedtest-go) - Protocol implementation reference
- [Speedtest.net](https://www.speedtest.net) - Service and API

---

**Document Version:** 2.0  
**Last Updated:** January 2026
