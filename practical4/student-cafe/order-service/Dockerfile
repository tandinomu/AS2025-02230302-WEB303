# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder
WORKDIR /app
# Copy go.mod and go.sum files
COPY go.mod go.sum ./
# Download all dependencies.
RUN go mod download
# Copy the source code
COPY . .
# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o /order-service .

# Stage 2: Create a minimal final image
FROM alpine:latest
WORKDIR /
# Copy the binary from the builder stage
COPY --from=builder /order-service /order-service
# Expose port 8081
EXPOSE 8081
# Command to run the executable
CMD ["/order-service"]
