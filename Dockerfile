# =========================
# Stage 1: Build
# =========================
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Download dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build API
RUN go build -o api ./cmd/api

# Build Worker
RUN go build -o worker ./cmd/worker


# =========================
# Stage 2: Run
# =========================
FROM alpine:latest

WORKDIR /app

# Copy both binaries
COPY --from=builder /app/api .
COPY --from=builder /app/worker .

# Run either API or Worker
# Docker Compose will override this command
CMD ["./api"]