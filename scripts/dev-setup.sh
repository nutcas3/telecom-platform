#!/bin/bash
set -e

echo "========================================="
echo "TaaS Platform - Development Setup"
echo "========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check Go 1.26
echo -n "Checking Go 1.26... "
if command_exists go; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ "$GO_VERSION" == 1.26* ]]; then
        echo -e "${GREEN}✓${NC} Found $(go version)"
    else
        echo -e "${YELLOW}⚠${NC} Found Go $GO_VERSION (1.26 recommended)"
    fi
else
    echo -e "${RED}✗${NC} Not found"
    echo "Please install Go 1.26: https://go.dev/dl/"
fi

# Check Rust 1.94
echo -n "Checking Rust 1.94... "
if command_exists rustc; then
    RUST_VERSION=$(rustc --version | awk '{print $2}')
    if [[ "$RUST_VERSION" == 1.94* ]]; then
        echo -e "${GREEN}✓${NC} Found $(rustc --version)"
    else
        echo -e "${YELLOW}⚠${NC} Found Rust $RUST_VERSION (1.94 recommended)"
    fi
else
    echo -e "${RED}✗${NC} Not found"
    echo "Please install Rust: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
fi

# Check Node.js 22
echo -n "Checking Node.js 22... "
if command_exists node; then
    NODE_VERSION=$(node --version | sed 's/v//')
    if [[ "$NODE_VERSION" == 22* ]]; then
        echo -e "${GREEN}✓${NC} Found $(node --version)"
    else
        echo -e "${YELLOW}⚠${NC} Found Node.js $NODE_VERSION (22 recommended)"
    fi
else
    echo -e "${RED}✗${NC} Not found"
    echo "Please install Node.js 22: https://nodejs.org/"
fi

# Check pnpm
echo -n "Checking pnpm... "
if command_exists pnpm; then
    echo -e "${GREEN}✓${NC} Found $(pnpm --version)"
else
    echo -e "${RED}✗${NC} Not found"
    echo "Installing pnpm..."
    npm install -g pnpm
fi

# Check MongoDB
echo -n "Checking MongoDB... "
if command_exists mongosh; then
    echo -e "${GREEN}✓${NC} Found"
else
    echo -e "${YELLOW}⚠${NC} MongoDB shell not found"
    echo "Please install MongoDB: https://www.mongodb.com/docs/manual/installation/"
fi

# Check Redis
echo -n "Checking Redis... "
if command_exists redis-cli; then
    echo -e "${GREEN}✓${NC} Found"
else
    echo -e "${YELLOW}⚠${NC} Redis not found"
    echo "Install Redis: sudo apt-get install redis-server"
fi

# Check Docker
echo -n "Checking Docker... "
if command_exists docker; then
    echo -e "${GREEN}✓${NC} Found $(docker --version | cut -d ' ' -f3)"
else
    echo -e "${YELLOW}⚠${NC} Docker not found (optional)"
fi

echo ""
echo "========================================="
echo "Installing Dependencies"
echo "========================================="
echo ""

# Install Go dependencies
echo "Installing Go dependencies..."
cd apps/api-server && go mod download && cd ../..
cd apps/carrier-connector && go mod download && cd ../..

# Install Rust dependencies
echo "Installing Rust dependencies..."
cargo fetch

# Install Node.js dependencies
echo "Installing Node.js dependencies..."
pnpm install

echo ""
echo "========================================="
echo "Creating Environment Files"
echo "========================================="
echo ""

# Create .env files from examples
for env_example in $(find . -name ".env.example"); do
    env_file="${env_example%.example}"
    if [ ! -f "$env_file" ]; then
        cp "$env_example" "$env_file"
        echo "Created $env_file"
    else
        echo "Skipped $env_file (already exists)"
    fi
done

echo ""
echo "========================================="
echo "Setup Complete!"
echo "========================================="
echo ""
echo "Next steps:"
echo "1. Set up databases: make db-setup"
echo "2. Build all components: make all"
echo "3. Start development: make dev"
echo ""
echo "For help: make help"
