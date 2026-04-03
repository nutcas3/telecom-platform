#!/bin/bash
set -e

echo "Setting up Redis for TaaS Platform..."

# Check if Redis is installed
if ! command -v redis-cli &> /dev/null; then
    echo "Redis CLI not found. Please install Redis first."
    echo "Ubuntu/Debian: sudo apt-get install redis-server"
    echo "macOS: brew install redis"
    exit 1
fi

# Redis connection details
REDIS_HOST=${REDIS_HOST:-"localhost"}
REDIS_PORT=${REDIS_PORT:-"6379"}

echo "Testing Redis connection..."

# Test connection
if redis-cli -h $REDIS_HOST -p $REDIS_PORT ping > /dev/null 2>&1; then
    echo "✓ Redis is running and accessible"
else
    echo "✗ Cannot connect to Redis at $REDIS_HOST:$REDIS_PORT"
    echo "Please start Redis: sudo systemctl start redis-server"
    exit 1
fi

# Configure Redis for TaaS
echo "Configuring Redis..."

redis-cli -h $REDIS_HOST -p $REDIS_PORT <<EOF
# Set memory limit
CONFIG SET maxmemory 2gb
CONFIG SET maxmemory-policy allkeys-lru

# Disable RDB persistence for speed (optional)
# CONFIG SET save ""

# Enable AOF for durability (optional)
# CONFIG SET appendonly yes

# Save configuration
CONFIG REWRITE
EOF

echo "Redis configuration complete!"
echo ""
echo "Connection string: redis://$REDIS_HOST:$REDIS_PORT/"
echo ""
echo "Update your .env files with:"
echo "REDIS_URI=redis://$REDIS_HOST:$REDIS_PORT/"

# Create some test data
echo ""
echo "Creating test credit balances..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT <<EOF
SET credit:10.0.0.1 1073741824
SET credit:10.0.0.2 5368709120
SET credit:10.0.0.3 10737418240
EOF

echo "✓ Test credit balances created"
echo "  - 10.0.0.1: 1 GB"
echo "  - 10.0.0.2: 5 GB"
echo "  - 10.0.0.3: 10 GB"
