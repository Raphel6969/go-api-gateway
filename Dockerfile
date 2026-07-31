# --- STAGE 1: Builder ---
FROM golang:1.26.4-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files and download them
COPY go.mod go.sum* ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build a statically linked binary for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway ./cmd/gateway

# --- STAGE 2: Production ---
FROM alpine:latest

WORKDIR /app

# Copy ONLY the compiled binary from the builder stage
COPY --from=builder /app/gateway .

# Expose the port the Gateway runs on
EXPOSE 8080

# Run the CLI tool with the start command
ENTRYPOINT ["./gateway", "start", "--config", "gateway.yaml"]
