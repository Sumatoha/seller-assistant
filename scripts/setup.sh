#!/bin/bash

# Seller Assistant Setup Script
# This script helps set up the development environment

set -e

echo "=================================="
echo "Seller Assistant Setup Script"
echo "=================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

echo "✓ Go is installed: $(go version)"

# Check if PostgreSQL is installed
if ! command -v psql &> /dev/null; then
    echo "⚠️  PostgreSQL client not found. Please install PostgreSQL."
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo "✓ PostgreSQL is installed"
fi

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "⚠️  Docker not found. Docker is recommended for deployment."
else
    echo "✓ Docker is installed"
fi

echo ""
echo "=================================="
echo "Setting up environment..."
echo "=================================="
echo ""

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file from template..."
    cp .env.example .env

    # Generate encryption key
    ENCRYPTION_KEY=$(openssl rand -base64 32)

    # Update .env with generated key
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s|ENCRYPTION_KEY=.*|ENCRYPTION_KEY=$ENCRYPTION_KEY|" .env
    else
        # Linux
        sed -i "s|ENCRYPTION_KEY=.*|ENCRYPTION_KEY=$ENCRYPTION_KEY|" .env
    fi

    echo "✓ .env file created with generated encryption key"
    echo ""
    echo "⚠️  IMPORTANT: Edit .env file and add your credentials:"
    echo "   - TELEGRAM_BOT_TOKEN"
    echo "   - OPENAI_API_KEY"
    echo "   - DATABASE_URL"
    echo ""
else
    echo "✓ .env file already exists"
fi

# Download dependencies
echo "Downloading Go dependencies..."
go mod download
echo "✓ Dependencies downloaded"

# Create bin directory
mkdir -p bin
echo "✓ Created bin directory"

echo ""
echo "=================================="
echo "Setup Complete!"
echo "=================================="
echo ""
echo "Next steps:"
echo ""
echo "1. Edit .env file with your credentials:"
echo "   nano .env"
echo ""
echo "2. Create and migrate database:"
echo "   createdb seller_assistant"
echo "   make migrate"
echo ""
echo "3. Run the application:"
echo "   make run-bot    (in one terminal)"
echo "   make run-worker (in another terminal)"
echo ""
echo "Or use Docker:"
echo "   make docker-up"
echo ""
echo "For more commands, run: make help"
echo ""
