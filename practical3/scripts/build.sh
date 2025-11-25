#!/bin/bash
set -e

echo "🚀 Building microservices..."

# Check dependencies
command -v docker >/dev/null 2>&1 || { echo "Docker required"; exit 1; }
command -v protoc >/dev/null 2>&1 || { echo "protoc required"; exit 1; }

# Generate proto files
echo "🔧 Generating proto files..."
protoc --go_out=./proto/gen --go_opt=paths=source_relative \
       --go-grpc_out=./proto/gen --go-grpc_opt=paths=source_relative \
       proto/*.proto

# Move generated files if needed
if [ -d "proto/gen/proto" ]; then
    mv proto/gen/proto/* proto/gen/
    rmdir proto/gen/proto
fi

# Distribute proto files to services
echo "📦 Distributing proto files..."
for service in api-gateway services/users-service services/products-service; do
    echo "  Copying to $service..."
    mkdir -p "$service/proto/gen"
    cp -r proto/gen/* "$service/proto/gen/"
done

# Clean up old containers
echo "🧹 Cleaning up..."
docker-compose down --remove-orphans 2>/dev/null || true

# Build and start
echo "🏗️  Building services..."
docker-compose build --no-cache

echo "🚀 Starting services..."
docker-compose up -d

echo "⏳ Waiting for services to be ready..."
sleep 30

echo ""
echo "✅ Services running at:"
echo "   - Consul UI: http://localhost:8500"
echo "   - API Gateway: http://localhost:8080"
echo "   - Users DB: localhost:5432"
echo "   - Products DB: localhost:5433"
echo ""
echo "📊 Check running containers:"
docker ps
