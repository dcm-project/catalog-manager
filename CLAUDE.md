# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DCM Catalog Manager is a Go-based service that uses OpenAPI specification-driven development. The project generates API types, server stubs, and client code from OpenAPI specifications located in `api/v1alpha1/openapi.yaml`.

## Build and Development Commands

```bash
# Build the binary
make build

# Run the application
make run

# Run tests
make test

# Run a single test
go test -run TestName ./path/to/package

# Format code
make fmt

# Vet code
make vet

# Clean build artifacts
make clean

# Tidy dependencies
make tidy
```

## Code Generation

This project uses `oapi-codegen` to generate Go code from OpenAPI specifications. After modifying `api/v1alpha1/openapi.yaml`, regenerate code:

```bash
# Regenerate all API-related code
make generate-api

# Or generate specific components:
make generate-types    # API models in api/v1alpha1/types.gen.go
make generate-spec     # Embedded spec in api/v1alpha1/spec.gen.go
make generate-server   # Chi server stubs in internal/api/server/server.gen.go
make generate-client   # Client code in pkg/client/client.gen.go

# Verify generated files are in sync
make check-generate-api
```

**Important**: Always run `make generate-api` after modifying the OpenAPI spec. The CI pipeline will fail if generated files are out of sync.

## API Standards Compliance

The project follows [AEP (API Enhancement Proposals)](https://aep.dev/) standards. OpenAPI specs are linted using Spectral:

```bash
# Check AEP compliance
make check-aep
```

The linter configuration is in `.spectral.yaml` and extends the AEP OpenAPI ruleset.

## Architecture

### Directory Structure

- **api/v1alpha1/**: OpenAPI specification and generation configs
  - `openapi.yaml`: Source of truth for API definitions (not yet created)
  - `*.gen.cfg`: oapi-codegen configuration files for different generation targets
  - Generated Go files (types, spec) will be placed here

- **cmd/catalog-manager/**: Main application entry point
  - `main.go`: Application bootstrap

- **internal/api/server/**: HTTP server implementation
  - `server.gen.cfg`: Generates Chi-based strict server interfaces
  - Generated server stubs use the Chi router with strict server pattern

- **pkg/client/**: Client library for consuming the API
  - `client.gen.cfg`: Generates client code that imports types from api/v1alpha1
  - Note: Client imports types from `github.com/dcm-project/policy-manager/api/v1alpha1` without namespace prefix

- **tools.go**: Build tools dependencies (oapi-codegen, ginkgo)

### Code Generation Configuration

Each `*.gen.cfg` file configures what oapi-codegen generates:

- **types.gen.cfg**: Generates data models only
- **spec.gen.cfg**: Generates embedded OpenAPI spec
- **server.gen.cfg**: Generates Chi server interfaces with strict server pattern
- **client.gen.cfg**: Generates HTTP client with custom imports

All configs use `skip-prune: true` to preserve manually written code.

## Testing

The project uses Ginkgo as the test framework. Test dependencies are managed via `tools.go`.

## CI/CD

GitHub Actions workflows enforce:

1. **CI** (`.github/workflows/ci.yaml`): Runs on all PRs to main
   - Build verification
   - Test execution

2. **Check Generated Files** (`.github/workflows/check-generate.yaml`): Runs when API files change
   - Ensures generated code is synchronized with OpenAPI spec

3. **Check AEP Compliance** (`.github/workflows/check-aep.yaml`): Runs when OpenAPI specs change
   - Validates API standards compliance
