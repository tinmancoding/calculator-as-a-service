# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies (none for this simple app using only stdlib)
RUN go mod download

# Copy source code
COPY main.go .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway .

# Final stage
FROM alpine:3.19

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Create non-root user for security
RUN adduser -D -u 1000 appuser
USER appuser

# Copy binary from builder stage
COPY --from=builder /app/gateway .

# Expose port
EXPOSE 8080

# Set environment variables
ENV PORT=8080
ENV SERVICE_NAME=gateway-service

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["./gateway"]
