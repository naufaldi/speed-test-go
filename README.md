# Speed Test CLI

A high-performance CLI tool written in Go that tests your internet connection speed (download and upload) and ping latency. This is a rewrite of the popular [sindresorhus/speed-test](https://github.com/sindresorhus/speed-test) Node.js CLI tool, implementing the speedtest.net protocol from scratch in pure Go.

## Features

- **Fast and Efficient**: Built in Go for optimal performance
- **Single Binary**: No runtime dependencies required
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Multiple Output Formats**: Human-readable, JSON, and verbose modes
- **Real-Time Progress**: Live speed updates during tests

## Installation

### Go Install

```bash
go install github.com/user/speed-test-go@latest
```

### Download Binary

Download pre-built binaries from the [releases page](https://github.com/user/speed-test-go/releases):

```bash
# Linux (amd64)
wget https://github.com/user/speed-test-go/releases/download/v1.0.0/speed-test-linux-amd64
chmod +x speed-test-linux-amd64
sudo mv speed-test-linux-amd64 /usr/local/bin/speed-test

# macOS (arm64)
curl -L https://github.com/user/speed-test-go/releases/download/v1.0.0/speed-test-darwin-arm64 -o speed-test
chmod +x speed-test
sudo mv speed-test /usr/local/bin/
```

### Homebrew (macOS/Linux)

```bash
brew install speed-test-go
```

### Docker

```bash
docker run --rm ghcr.io/user/speed-test-go
```

## Usage

### Basic Speed Test

```bash
$ speed-test
      Ping 24.5 ms
  Download 95.32 Mbps
    Upload 23.45 Mbps
```

### JSON Output

```bash
$ speed-test --json
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
  }
}
```

### Verbose Output

```bash
$ speed-test --verbose
      Ping 24.5 ms
  Download 95.32 Mbps
    Upload 23.45 Mbps

    Server   speedtest.server.com
  Location New York (United States)
  Distance 1,245.3 km
```

### Specify Server

```bash
$ speed-test --server 1234
```

### Options

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output the result as JSON |
| `--bytes` | `-b` | Output in megabytes per second (MBps) |
| `--verbose` | `-v` | Output detailed information including server details |
| `--server` | `-s` | Specify a server ID to use for testing |
| `--timeout` | `-t` | Timeout for the speed test (default: 30s) |
| `--help` | `-h` | Show help information |
| `--version` | `-V` | Show version information |

### Help

```bash
$ speed-test --help
Test your internet connection speed and ping using speedtest.net from the CLI.
    
Supports multiple output formats and configuration options.

Usage:
  speed-test [flags]

Flags:
  -b, --bytes              Output the result in megabytes per second (MBps)
  -h, --help               help for speed-test
  -j, --json               Output the result as JSON
  -s, --server string      Specify a server ID to use
  -t, --timeout duration   Timeout for the speed test (default 30s)
  -v, --verbose            Output more detailed information
```

## Building

### Prerequisites

- Go 1.21 or later
- Make

### Build

```bash
make build
```

This creates the `speed-test` binary in the current directory.

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

### All Targets

```bash
make help
```

## Development

### Project Structure

```
speed-test-go/
├── cmd/                    # CLI commands
│   ├── root.go            # Main command
│   └── version.go         # Version command
├── internal/              # Internal packages
│   ├── location/         # User location detection
│   ├── network/          # HTTP client
│   ├── output/           # Output formatting
│   ├── server/           # Server discovery
│   ├── test/             # Test runner
│   └── transfer/         # Download/upload tests
├── pkg/                   # Public packages
│   └── types/            # Type definitions
├── .github/workflows/    # CI/CD
├── docs/                  # Documentation
├── main.go               # Entry point
├── Makefile              # Build automation
└── README.md             # This file
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Running Tests

```bash
# Run all tests with coverage
make test

# Run specific package tests
go test -v ./internal/...

# Run with race detector
go test -race ./...
```

### Code Style

This project uses:
- [golangci-lint](https://golangci-lint.run/) for linting
- [gofumpt](https://github.com/mvdan/gofumpt) for formatting
- Standard Go conventions

```bash
# Format code
gofumpt -w .

# Run linters
make lint
```

## Architecture

### Test Flow

1. **Server Discovery**: Fetch list of speed test servers from speedtest.net
2. **User Location**: Detect user's IP and geographic location
3. **Server Selection**: Calculate distances and select nearest server
4. **Ping Test**: Measure latency to selected server (5 requests)
5. **Download Test**: Measure download bandwidth (multi-threaded)
6. **Upload Test**: Measure upload bandwidth (multi-threaded)

### Output Formats

The CLI supports three output modes:

1. **Human-Readable**: Default format matching sindresorhus/speed-test
2. **JSON**: Structured output for scripting (`--json`)
3. **Verbose**: Detailed output with server information (`--verbose`)

## Performance

- **Binary Size**: < 10MB
- **Memory Usage**: < 50MB
- **Test Duration**: < 30 seconds
- **Code Coverage**: > 80%

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [sindresorhus/speed-test](https://github.com/sindresorhus/speed-test) - Original Node.js CLI inspiration
- [showwin/speedtest-go](https://github.com/showwin/speedtest-go) - Protocol implementation reference
- [speedtest.net](https://www.speedtest.net) - Speed test service and API
