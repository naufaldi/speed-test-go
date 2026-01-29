# AGENTS.md

## Commands
- **Build**: `make build` or `go build -o speed-test main.go`
- **Test all**: `make test` or `go test -v -cover ./...`
- **Test single**: `go test -v -run TestName ./path/to/package`
- **Lint**: `make lint` or `golangci-lint run ./...`
- **Coverage**: `make test-coverage`

## Architecture
- **cmd/**: CLI commands using Cobra (root.go, version.go)
- **internal/**: Core logic (location, network, output, server, test, transfer)
- **pkg/types/**: Shared type definitions
- **main.go**: Entry point, delegates to cmd.Execute()

## Code Style
- Use `gofmt` and `goimports` for formatting
- Import order: stdlib, external, local (`github.com/user/speed-test-go`)
- Always check and handle errors explicitly
- Keep internal packages private; expose via pkg/ for shared types
- Follow standard Go naming: camelCase for private, PascalCase for exported
