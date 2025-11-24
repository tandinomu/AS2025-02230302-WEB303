FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copy go mod files
COPY time-service/go.mod time-service/go.sum ./

# Copy proto files
COPY proto/ ./proto/

# Download dependencies
RUN go mod download

# Copy service code
COPY time-service/ ./time-service/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./time-service/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 50052
CMD ["/app/server"]
