# DevOps Specification (Delta)

## Purpose

Define Docker and build automation requirements for MCP Policy Proxy.

## ADDED Requirements

### Requirement: Docker Build - Multi-Stage Build

The Dockerfile SHALL use a multi-stage build to minimize image size and attack surface.

#### Dockerfile Requirements

**Stage 1: Builder**
- Base image: `golang:1.21-alpine`
- Install git for module downloads
- Download Go modules
- Build binary with `CGO_ENABLED=0` (static binary)
- Strip symbols and disable DWARF: `-ldflags="-s -w"`

**Stage 2: Runtime**
- Base image: `alpine:3.19`
- Install `ca-certificates` for HTTPS
- Copy only the binary from builder stage
- Set non-root user

### Requirement: Docker Security Hardening

The container SHALL run as non-root user with minimal privileges.

**User Requirements:**
- Create user `appuser` with UID 1000
- Set working directory ownership to `appuser`
- Use `USER appuser` directive before CMD
- Binary MUST be readable and executable by non-root

**Filesystem Requirements:**
- No world-writable files
- No sensitive files in image (credentials, keys)
- Use `.dockerignore` to exclude source files and test files

#### Scenario: Container runs as non-root

- GIVEN Docker image is built
- WHEN container is started
- THEN process SHALL run as UID 1000
- AND `ps aux` SHALL show `appuser`
- AND `/proc/self/uid_map` SHALL show non-zero mapped root

### Requirement: Dockerfile Build Arguments

The Dockerfile SHALL accept build arguments for versioning.

**Build Arguments:**
- `VERSION`: Git tag or build number (default: `dev`)
- `BUILD_DATE`: ISO 8601 timestamp of build

**Labels:**
- `org.opencontainers.image.version` = VERSION
- `org.opencontainers.image.created` = BUILD_DATE
- `org.opencontainers.image.title` = "MCP Policy Proxy"

#### Scenario: Build with version argument

- GIVEN Docker build command with `--build-arg VERSION=v1.2.3`
- WHEN image is built
- THEN `docker inspect` SHALL show label `org.opencontainers.image.version=v1.2.3`

### Requirement: Makefile Targets

The project root SHALL contain a Makefile with standardized targets.

#### Makefile Targets

| Target | Description |
|--------|-------------|
| `build` | Compile Go binary for current platform |
| `test` | Run unit tests with coverage |
| `lint` | Run golangci-lint if available |
| `docker-build` | Build Docker image with version tag |
| `docker-run` | Run Docker container with defaults |
| `docker-stop` | Stop running container |
| `clean` | Remove binaries and build artifacts |

**Build Target:**
- Output: `mcp-policy-proxy` (or `mcp-policy-proxy.exe` on Windows)
- Use `go build` with version ldflags

**Test Target:**
- Run `go test -v -coverprofile=coverage.out`
- Generate HTML coverage: `go tool cover -html=coverage.out`

**Docker Build Target:**
- Build image tagged `hexstrike/mcp-policy-proxy:latest` and `hexstrike/mcp-policy-proxy:<version>`
- Pass VERSION and BUILD_DATE arguments

**Docker Run Target:**
- Run container with port mapping `8080:8080`
- Set required environment variables
- Use `hexstrike/mcp-policy-proxy:latest`

#### Scenario: Build on local machine

- GIVEN Go 1.21+ installed
- WHEN `make build` is executed
- THEN binary `mcp-policy-proxy` SHALL be created
- AND `make test` SHALL run all unit tests

#### Scenario: Docker build with custom version

- GIVEN Docker installed
- WHEN `make docker-build VERSION=2.0.0` is executed
- THEN image SHALL be tagged `hexstrike/mcp-policy-proxy:2.0.0`
- AND `docker images` SHALL show the new image

#### Scenario: Clean removes artifacts

- GIVEN previous build artifacts exist
- WHEN `make clean` is executed
- THEN `mcp-policy-proxy` binary SHALL be deleted
- AND `coverage.out` SHALL be deleted

### Requirement: .dockerignore

The project SHALL include `.dockerignore` to minimize build context.

**Required Exclusions:**
```
.git
.gitignore
*.md
docs/
openspec/
tests/
*_test.go
Makefile
.env
*.log
coverage.out
```

#### Scenario: Build context excludes unnecessary files

- GIVEN `.dockerignore` exists with exclusions
- WHEN Docker build is executed
- THEN excluded files SHALL NOT be sent to Docker daemon
- AND build context size SHALL be minimal
