# AgentBox Server Dockerfile
# Multi-stage build for minimal image size

# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o agentbox ./cmd/agentbox

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk add --no-cache ca-certificates docker-cli

# Copy binary from builder
COPY --from=builder /build/agentbox /app/agentbox

# Create workspace directory
RUN mkdir -p /tmp/agentbox/workspaces

# Environment variables
ENV AGENTBOX_HOST=0.0.0.0
ENV AGENTBOX_PORT=8080
ENV AGENTBOX_WORKSPACE_BASE=/tmp/agentbox/workspaces

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

# Run
ENTRYPOINT ["/app/agentbox"]
