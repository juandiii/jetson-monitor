## Multi-stage Dockerfile with module-aware build and minimal runtime
## Build stage
ARG GOVERSION=1.25.4
FROM golang:${GOVERSION}-alpine AS builder
WORKDIR /src

# Install tools needed for module downloads
RUN apk add --no-cache git ca-certificates

# Cache modules by copying go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the sources and build a statically linked binary
COPY . .
ARG GOOS=linux
ARG GOARCH=amd64
ARG CGO_ENABLED=0
ARG LDFLAGS="-s -w"
# Build only the main package in repository root to produce a single binary
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -o /out/jetson-monitor -ldflags="${LDFLAGS}" .

## Runtime stage: small base image with CA certs and non-root user
FROM alpine:3.18 AS runtime
RUN apk add --no-cache ca-certificates curl

# non-root user for improved security
RUN addgroup -S jetgroup && adduser -S -G jetgroup -h /home/jetson -s /sbin/nologin -u 1000 jetson

WORKDIR /home/jetson
COPY --from=builder /out/jetson-monitor /usr/local/bin/jetson-monitor
RUN chmod +x /usr/local/bin/jetson-monitor

USER jetson
EXPOSE 38080

# Healthcheck uses the application's /health endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s CMD curl -fsS http://localhost:38080/health || exit 1

ENTRYPOINT ["/usr/local/bin/jetson-monitor"]
CMD []
