#!/bin/bash

# Comprehensive deployment verification script
echo "=== Telecom Platform Deployment Verification ==="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check service status
check_service() {
    local service=$1
    local container=$2
    echo -n "Checking $service... "
    
    if docker ps --format "table {{.Names}}" | grep -q "$container"; then
        echo -e "${GREEN}RUNNING${NC}"
        return 0
    else
        echo -e "${RED}NOT RUNNING${NC}"
        return 1
    fi
}

# Function to check port availability
check_port() {
    local port=$1
    local service=$2
    echo -n "Checking $service port $port... "
    
    if netstat -tuln 2>/dev/null | grep -q ":$port "; then
        echo -e "${GREEN}OPEN${NC}"
        return 0
    else
        echo -e "${RED}CLOSED${NC}"
        return 1
    fi
}

# Function to check configuration files
check_config() {
    local file=$1
    local service=$2
    echo -n "Checking $service config... "
    
    if [[ -f "$file" ]]; then
        echo -e "${GREEN}EXISTS${NC}"
        return 0
    else
        echo -e "${RED}MISSING${NC}"
        return 1
    fi
}

echo ""
echo "=== Docker Services Status ==="
check_service "MongoDB (App)" "taas-mongodb"
check_service "Redis" "taas-redis"
check_service "MongoDB (free5GC)" "free5gc-mongo"
check_service "NRF" "nrf"
check_service "AMF" "amf"
check_service "SMF" "smf"
check_service "UPF" "upf"
check_service "UDR" "udr"
check_service "UDM" "udm"
check_service "AUSF" "ausf"
check_service "NSSF" "nssf"
check_service "PCF" "pcf"
check_service "API Server" "taas-api-server"
check_service "Carrier Connector" "taas-carrier-connector"
check_service "Charging Engine" "taas-charging-engine"
check_service "Web Dashboard" "taas-web-dashboard"

echo ""
echo "=== Port Availability ==="
check_port "27017" "MongoDB (App)"
check_port "6379" "Redis"
check_port "8000" "API Server"
check_port "8080" "Charging Engine"
check_port "3000" "Web Dashboard"

echo ""
echo "=== Configuration Files ==="
check_config "deployments/free5gc/config/nrfcfg.yaml" "NRF"
check_config "deployments/free5gc/config/amfcfg.yaml" "AMF"
check_config "deployments/free5gc/config/smfcfg.yaml" "SMF"
check_config "deployments/free5gc/config/upfcfg.yaml" "UPF"
check_config "deployments/free5gc/config/udrcfg.yaml" "UDR"
check_config "deployments/free5gc/config/udmcfg.yaml" "UDM"
check_config "deployments/free5gc/config/ausfcfg.yaml" "AUSF"
check_config "deployments/free5gc/config/nssfcfg.yaml" "NSSF"
check_config "deployments/free5gc/config/pcfcfg.yaml" "PCF"

echo ""
echo "=== Network Connectivity ==="
echo -n "Testing MongoDB connection... "
if docker exec taas-mongo mongosh --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED${NC}"
fi

echo -n "Testing Redis connection... "
if docker exec taas-redis redis-cli ping >/dev/null 2>&1; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED${NC}"
fi

echo ""
echo "=== Summary ==="
echo "To start all services: docker-compose up -d"
echo "To start free5GC only: make free5gc-start"
echo "To view logs: make free5gc-logs"
echo "To stop all services: docker-compose down"
echo ""
echo "For detailed service logs:"
echo "  docker-compose logs -f [service-name]"
echo ""
echo "Web interfaces:"
echo "  - Web Dashboard: http://localhost:3000"
echo "  - API Server: http://localhost:8000"
echo "  - Charging Engine: http://localhost:8080"
