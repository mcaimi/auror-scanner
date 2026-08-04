# AURor Scanner container builder

# ============================================================================
# Stage 1: Build Go binary
# ============================================================================
FROM golang:1.26-alpine AS go-builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN make all

# ============================================================================
# Stage 2: Minimal runtime image
# ============================================================================
FROM alpine:latest

# Create non-root user
RUN addgroup -g 1000 auror && \
  adduser -D -u 1000 -G auror auror

# Create directories for data
RUN mkdir -p /config && \
  chown -R auror:auror /config && \
  mkdir -p /app/workspace && \
  chown -R auror:auror /app/workspace

WORKDIR /app

# Copy binary from builder
COPY --from=go-builder /build/build/auror /app/auror

# Copy default config
COPY --from=go-builder /build/auror.yaml /config/auror.yaml

# Copy prompt and skills
COPY --from=go-builder /build/prompts /config/prompts
COPY --from=go-builder /build/skill /config/skill

# Set ownership
RUN chown -R auror:auror /app && chown -R auror:auror /config

# Switch to non-root user
USER auror

# Volume for persistent data
VOLUME ["/config"]
VOLUME ["/app/workspace"]

# Default command
ENTRYPOINT ["/app/auror", "--config", "/config/auror.yaml"]
CMD [""]