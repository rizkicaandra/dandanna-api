# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

# Copy go.mod first for better layer caching.
# go.sum is only present once external dependencies are added — copy it if it exists.
COPY go.mod ./
COPY go.sum* ./

RUN go mod download

COPY . .

RUN VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo dev) && \
    BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
        -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
        -o dandanna-api ./cmd/api

# Final stage — pin to a specific alpine version, never use :latest
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/dandanna-api .

ENV TZ=UTC

# Drop root — run as unprivileged user
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/api/health || exit 1

CMD ["./dandanna-api"]
