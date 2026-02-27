# Build stage
FROM registry.access.redhat.com/ubi9/go-toolset:1.25.5 AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
USER root
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o catalog-manager ./cmd/catalog-manager

# Runtime stage
FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
WORKDIR /app
COPY --from=builder /app/catalog-manager .
EXPOSE 8080

# Use SQLite for easy local testing (no external DB required)
ENV DB_TYPE=sqlite
ENV DB_NAME=/tmp/catalog.db

ENTRYPOINT ["./catalog-manager"]
