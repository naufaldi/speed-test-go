# Release v1.0.0

## Features
- High-performance CLI speed test tool in pure Go
- Implements speedtest.net protocol from scratch
- Multi-threaded download/upload tests
- Ping-based server selection
- Multiple output formats (human-readable, JSON, verbose)
- Cross-platform support (Linux, macOS, Windows)

## CLI Options
| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output as JSON |
| `--bytes` | `-b` | Output in MB/s |
| `--verbose` | `-v` | Detailed server info |
| `--server` | `-s` | Specify server ID |
| `--servers` | `-n` | Test N closest servers (default: 5) |
| `--timeout` | `-t` | Test timeout (default: 30s) |
| `--version` | `-V` | Show version |

## Installation
```bash
# Go install
go install github.com/user/speed-test-go@latest

# Or download from releases
```

## Binary Checksums (SHA256)
- `speed-test-darwin-amd64`: ...
- `speed-test-darwin-arm64`: ...
- `speed-test-linux-amd64`: ...
- `speed-test-linux-arm64`: ...
- `speed-test-windows-amd64.exe`: ...
