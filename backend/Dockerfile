# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git (needed for go mod download)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o feedback ./cmd

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/feedback .

# Copy locales
COPY --from=builder /app/locales ./locales

# Create logs directory
RUN mkdir -p logs

# Expose port
EXPOSE 5012

# Run the application
CMD ["./feedback"]
