# Multi-stage Dockerfile for Paxeer Launchpad
# Builds both the Go indexer/API and prepares for frontend deployment

# Stage 1: Build Go services
FROM golang:1.22-alpine AS go-builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy Go modules
COPY Off-Chain-Server/go.mod Off-Chain-Server/go.sum ./Off-Chain-Server/
WORKDIR /app/Off-Chain-Server

# Download dependencies
RUN go mod download

# Copy source code
COPY Off-Chain-Server/ .

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/indexer ./cmd/indexer
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api ./cmd/api

# Stage 2: Node.js for smart contracts and frontend
FROM node:18-alpine AS node-builder

# Install pnpm
RUN npm install -g pnpm

# Set working directory
WORKDIR /app

# Copy package files
COPY Smart-Contracts/package.json Smart-Contracts/pnpm-lock.yaml ./Smart-Contracts/
WORKDIR /app/Smart-Contracts

# Install dependencies
RUN pnpm install --frozen-lockfile

# Copy smart contract source
COPY Smart-Contracts/ .

# Compile contracts and export ABIs
RUN pnpm hardhat compile
RUN pnpm hardhat run scripts/export_abis.js

# Stage 3: Production runtime
FROM alpine:3.18 AS runtime

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata postgresql-client curl

# Create app user
RUN addgroup -g 1001 -S paxeer && \
    adduser -S paxeer -u 1001 -G paxeer

# Set working directory
WORKDIR /app

# Copy Go binaries from builder
COPY --from=go-builder /app/Off-Chain-Server/bin/ ./bin/
COPY --from=go-builder /app/Off-Chain-Server/configs/ ./configs/
COPY --from=go-builder /app/Off-Chain-Server/migrations/ ./migrations/

# Copy ABI exports
COPY --from=node-builder /app/Smart-Contracts/abi-dist/ ./abi-dist/

# Create logs directory
RUN mkdir -p logs && chown -R paxeer:paxeer /app

# Switch to app user
USER paxeer

# Expose API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Default command (can be overridden)
CMD ["./bin/api"]