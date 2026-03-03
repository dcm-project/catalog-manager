# Build stage
FROM registry.access.redhat.com/ubi9/go-toolset:1.25.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

USER root
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o catalog-manager ./cmd/catalog-manager

# Runtime stage
FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

WORKDIR /app

COPY --from=builder /app/catalog-manager .

EXPOSE 8080

# DB configuration is provided via environment variables at runtime
# (e.g., via docker-compose, Kubernetes manifests, or make run)
# For local dev: make run sets DB_TYPE=sqlite DB_NAME=/tmp/catalog.db
# SQLite: set DB_TYPE=sqlite and DB_NAME to path (e.g. /tmp/catalog.db)

ENTRYPOINT ["./catalog-manager"]
