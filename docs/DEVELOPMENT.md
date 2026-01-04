# Development Guide

## Prerequisites

- Go 1.25.5 or higher
- Gomobile (for mobile development)
- Git

## Building the Project

### Desktop/Server Application

```bash
# Build the main application
go build -o bin/intermesh ./cmd/intermesh

# Run with default settings
./bin/intermesh

# Run with custom settings
./bin/intermesh -id="node-1" -name="My Node" -ip="192.168.1.100" -internet=true
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestNodeCreation ./pkg/mesh
```

### Mobile Development

#### iOS

```bash
# Install Gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

# Build for iOS
gomobile bind -target=ios -o intermesh.xcframework ./mobile
```

#### Android

```bash
# Build for Android
gomobile bind -target=android -o intermesh.aar ./mobile
```

## Project Structure

```
intermesh/
├── cmd/
│   └── intermesh/       # Main CLI application
├── pkg/
│   └── mesh/            # Core mesh networking logic
│       ├── node.go      # Node and peer management
│       ├── network.go   # Personal networks
│       ├── routing.go   # Routing logic
│       ├── proxy.go     # Proxy management
│       └── mesh_test.go # Unit tests
├── internal/            # Internal packages (to be added)
├── mobile/              # Mobile platform bindings
│   └── intermesh.go     # Gomobile wrapper
├── docs/                # Documentation
├── test/                # Integration tests (to be added)
├── go.mod               # Go module definition
└── README.md            # Project README
```

## Code Organization

### Package: `pkg/mesh`

Core mesh networking library with:
- **Node Management**: Node and Peer structures
- **Personal Networks**: Sub-mesh groups and policies
- **Routing**: Route selection and management
- **Proxy Management**: Internet sharing via proxies

### Package: `mobile`

Mobile-specific bindings that wrap mesh package for use on:
- iOS via Gomobile
- Android via Gomobile

### Command: `cmd/intermesh`

CLI application for:
- Starting a mesh node
- Managing node configuration
- Monitoring network status

## Development Workflow

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/mesh-discovery
   ```

2. **Implement Changes**
   - Write code in appropriate package
   - Add tests alongside implementation

3. **Run Tests**
   ```bash
   go test ./...
   ```

4. **Commit and Push**
   ```bash
   git add .
   git commit -m "feat: implement peer discovery"
   git push origin feature/mesh-discovery
   ```

5. **Create Pull Request**
   - Describe changes
   - Reference related issues

## Common Tasks

### Adding a New Feature to the Mesh Package

1. Create a new file: `pkg/mesh/feature.go`
2. Implement the feature with clear interfaces
3. Add tests in: `pkg/mesh/mesh_test.go` or create `pkg/mesh/feature_test.go`
4. Update documentation if necessary

### Adding Mobile Bindings

1. Add wrapper in `mobile/intermesh.go`
2. Ensure all public functions are capitalized (Go naming convention)
3. Test with Gomobile binding

### Adding Internal Utilities

1. Create package in `internal/`
2. Use for internal-only functionality
3. Export only what's necessary

## Debugging

### Enable Logging

The main application logs important events. For more detail, consider adding:

```go
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

### Run with Debug Output

```bash
go run -v ./cmd/intermesh
```

### Unit Test Debugging

```bash
go test -v -run TestNodeCreation ./pkg/mesh
```

## Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./pkg/mesh
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./pkg/mesh
go tool pprof mem.prof
```

## Dependencies

The project aims to be lightweight with minimal dependencies. Current core dependencies:
- Standard library only for now
- Gomobile for mobile platform bindings

## Contributing Guidelines

1. Follow Go conventions (gofmt, golint)
2. Write tests for new functionality
3. Update documentation for significant changes
4. Keep commits logical and descriptive
