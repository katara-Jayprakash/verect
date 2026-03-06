# Stage 1: Build the Go binary
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Copy and install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o reverse-go-proxy .

# Production Image
FROM alpine:latest

# S3 REQUIRES certificates to connect over HTTPS
RUN apk --no-cache add ca-certificates

# Set a clean directory for the app
WORKDIR /root/

COPY --from=builder /app/reverse-go-proxy .

#  listens on Port 80
EXPOSE 80
CMD ["./reverse-go-proxy"]
