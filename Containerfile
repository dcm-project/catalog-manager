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

ENTRYPOINT ["./catalog-manager"]
