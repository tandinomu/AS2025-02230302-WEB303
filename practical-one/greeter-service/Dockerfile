FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copy go mod files
COPY greeter-service/go.mod greeter-service/go.sum ./

# Copy proto files
COPY proto/ ./proto/

# Download dependencies
RUN go mod download

# Copy service code
COPY greeter-service/ ./greeter-service/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./greeter-service/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 50051
CMD ["/app/server"]
