# DCM Catalog Manager

DCM Catalog Manager is a Go service for managing service type definitions,
catalog items, and catalog item instances for infrastructure services. It
provides a RESTful API following
[AEP (API Enhancement Proposals)](https://aep.dev/) standards with an OpenAPI
specification-driven development workflow.

## Overview

The Catalog Manager enables a hierarchical resource model for defining and
provisioning infrastructure services:

```
ServiceType → CatalogItem → CatalogItemInstance
```

- **ServiceTypes** define templates for infrastructure services (VM, container,
  database, Kubernetes cluster, or custom types) with an opaque spec schema.
- **CatalogItems** are curated service offerings built on top of a ServiceType.
  They define field configurations with defaults, editability constraints,
  validation schemas, and conditional dependencies (`depends_on`).
- **CatalogItemInstances** represent concrete requests to provision a service.
  On creation, the system resolves the full reference chain (Instance →
  CatalogItem → ServiceType), constructs a resource spec by merging defaults and
  user values, validates inputs, and forwards the spec to the Placement Manager
  for provisioning.

## Architecture

The service follows a three-layer architecture:

```
OpenAPI Validation Middleware → Handler Layer (HTTP) → Service Layer (Business Logic) → Store Layer (Database)
```

- **OpenAPI Middleware** — Validates all incoming requests against the OpenAPI
  spec (required fields, types, formats, patterns) before they reach handlers.
- **Handler Layer** (`internal/handlers/v1alpha1/`) — Thin HTTP handlers that
  delegate to the service layer and map domain errors to HTTP responses.
- **Service Layer** (`internal/service/`) — Business logic, validation, ID/path
  generation, model conversion, spec construction, and Placement Manager
  integration.
- **Store Layer** (`internal/store/`) — GORM-based persistence with PostgreSQL
  and SQLite support, pagination, and foreign key constraint enforcement.

### Directory Structure

```
cmd/catalog-manager/          Main application entry point
api/v1alpha1/
  openapi.yaml                Source of truth for the API
  types.gen.go                Generated API types
  spec.gen.go                 Generated embedded spec
  servicetypes/               Service type spec definitions and generated types
    common.yaml               Common fields shared across service types
    vm/                       Virtual Machine spec
    container/                Container spec
    database/                 Database spec
    cluster/                  Kubernetes Cluster spec
    three_tier_app_demo/      Three-Tier App Demo spec
internal/
  api/server/                 Generated Chi server stubs
  apiserver/                  Server setup and middleware
  config/                     Configuration (env-based)
  handlers/v1alpha1/          HTTP handlers and error mapping
  placement/                  Placement Manager client
  service/                    Business logic, spec builder, seed data
  store/                      GORM store implementation
    model/                    Database models
pkg/client/                   Generated Go client library
```

## API

The API is served at `/api/v1alpha1` and follows AEP resource-oriented design
with RFC 7807 error responses.

### Endpoints

| Method   | Path                           | Description                                                  |
| -------- | ------------------------------ | ------------------------------------------------------------ |
| `GET`    | `/health`                      | Health check                                                 |
| `GET`    | `/service-types`               | List service types (paginated)                               |
| `POST`   | `/service-types`               | Create a service type                                        |
| `GET`    | `/service-types/{id}`          | Get a service type                                           |
| `GET`    | `/catalog-items`               | List catalog items (paginated, filterable by `service_type`) |
| `POST`   | `/catalog-items`               | Create a catalog item                                        |
| `GET`    | `/catalog-items/{id}`          | Get a catalog item                                           |
| `PATCH`  | `/catalog-items/{id}`          | Update a catalog item (JSON Merge Patch)                     |
| `DELETE` | `/catalog-items/{id}`          | Delete a catalog item                                        |
| `GET`    | `/catalog-item-instances`      | List instances (paginated, filterable by `catalog_item_id`)  |
| `POST`   | `/catalog-item-instances`      | Create an instance                                           |
| `GET`    | `/catalog-item-instances/{id}` | Get an instance                                              |
| `DELETE` | `/catalog-item-instances/{id}` | Delete an instance                                           |

### Resource IDs

All resources support optional user-specified IDs via the `id` query parameter
(DNS-1123 format: lowercase alphanumeric with hyphens, max 63 characters). If
omitted, a UUID is generated automatically.

### Key Features

- **Pagination** — Token-based pagination with configurable page size (default:
  100, max: 1000).
- **Field Configurations** — CatalogItems define fields with `path`, `editable`,
  `default`, `validation_schema` (JSON Schema), and `depends_on` (conditional
  defaults/options).
- **Dependency Validation** — Cyclic `depends_on` references are rejected at
  CatalogItem creation. User values are validated against allowed options at
  instance creation.
- **Immutability** — `api_version` and `spec.service_type` are immutable after
  CatalogItem creation.
- **Foreign Key Constraints** — Database-level referential integrity
  (ServiceType → CatalogItem → CatalogItemInstance) with `RESTRICT` on delete.
- **Spec Construction** — On instance creation, the spec builder resolves the
  full reference chain: looks up the CatalogItem, resolves its ServiceType, then
  deep-copies the ServiceType spec as a base template. It applies CatalogItem
  field defaults, then overlays user-provided values — validating that each
  targets a known, editable field and passes its JSON Schema constraint.
  Finally, it validates `depends_on` constraints against the fully assembled
  spec.

## Getting Started

### Prerequisites

- Go 1.25.5+
- PostgreSQL (production) or SQLite (development)
- [Spectral](https://stoplight.io/open-source/spectral) (for AEP linting)

### Configuration

Configuration is loaded from environment variables:

| Variable                | Default                 | Description                             |
| ----------------------- | ----------------------- | --------------------------------------- |
| `BIND_ADDRESS`          | `0.0.0.0:8080`          | HTTP server bind address                |
| `DB_TYPE`               | `pgsql`                 | Database type (`pgsql` or `sqlite`)     |
| `DB_HOST`               | `localhost`             | PostgreSQL host                         |
| `DB_PORT`               | `5432`                  | PostgreSQL port                         |
| `DB_NAME`               | `catalog-manager`       | Database name (or file path for SQLite) |
| `DB_USER`               | `admin`                 | PostgreSQL user                         |
| `DB_PASSWORD`           | `adminpass`             | PostgreSQL password                     |
| `PLACEMENT_MANAGER_URL` | `http://localhost:8081` | Placement Manager base URL              |

### Build and Run

```bash
# Build the binary
make build

# Run locally with SQLite
make run

# Run with PostgreSQL
DB_TYPE=pgsql DB_HOST=localhost DB_NAME=catalog-manager go run ./cmd/catalog-manager
```

The service auto-migrates the database schema on startup and seeds default
service types and catalog items (e.g., the "Pet Clinic" three-tier app demo).

## Development

### Common Commands

```bash
make build              # Build the binary
make run                # Run with SQLite (development)
make test               # Run all tests (Ginkgo)
make fmt                # Format code
make vet                # Vet code
make tidy               # Tidy dependencies
make clean              # Clean build artifacts
```

### Running a Single Test

```bash
go test -run TestName ./path/to/package
```

### Code Generation

The project uses `oapi-codegen` to generate Go code from OpenAPI specifications.
After modifying `api/v1alpha1/openapi.yaml`:

```bash
# Regenerate all API code (types, spec, server, client, service types)
make generate-api

# Or generate specific components
make generate-types          # API models
make generate-spec           # Embedded spec
make generate-server         # Chi server stubs
make generate-client         # Client library
make generate-service-types  # Service type specs (VM, Container, Database, Cluster, etc.)

# Verify generated files are in sync
make check-generate-api
```

### AEP Compliance

```bash
make check-aep
```

Lints the OpenAPI spec against AEP standards using Spectral (config in
`.spectral.yaml`).

### Testing

The project uses [Ginkgo](https://onsi.github.io/ginkgo/) as the test framework
with [Gomega](https://onsi.github.io/gomega/) matchers. Tests use in-memory
SQLite databases and mock interfaces for isolated unit testing.

Test suites cover:

- **Handler tests** — HTTP request/response mapping with mock services
- **Service tests** — Business logic with mock stores
- **Store tests** — Database operations with in-memory SQLite
- **Integration tests** — Full hierarchy creation, FK constraints, deletion
  workflows
- **Placement client tests** — HTTP client behavior with test servers

## CI/CD

GitHub Actions workflows enforce:

1. **CI** (`.github/workflows/ci.yaml`) — Build and test on all PRs to main.
2. **Check Generated Files** (`.github/workflows/check-generate.yaml`) — Ensures
   generated code is in sync with the OpenAPI spec.
3. **Check AEP Compliance** (`.github/workflows/check-aep.yaml`) — Validates API
   standards compliance.

## License

Apache 2.0 — see the [OpenAPI spec](api/v1alpha1/openapi.yaml) for license
details.
