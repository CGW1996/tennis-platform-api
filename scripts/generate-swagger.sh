#!/bin/bash

# Generate Swagger documentation for Tennis Platform API
# This script generates swagger.json, swagger.yaml, and docs.go files

echo "Generating Swagger documentation..."

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo "swag command not found. Installing..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Generate Swagger docs
swag init -g cmd/server/main.go -o docs

if [ $? -eq 0 ]; then
    echo "✅ Swagger documentation generated successfully!"
    echo "📁 Files generated:"
    echo "   - docs/docs.go"
    echo "   - docs/swagger.json"
    echo "   - docs/swagger.yaml"
    echo ""
    echo "🌐 Access Swagger UI at: http://localhost:8080/swagger/index.html"
    echo "📋 API JSON at: http://localhost:8080/swagger/doc.json"
else
    echo "❌ Failed to generate Swagger documentation"
    exit 1
fi