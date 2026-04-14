#!/bin/bash

# Test script for free5GC deployment
echo "=== Testing free5GC Configuration ==="

# Check if configuration files exist
echo "Checking configuration files..."
config_files=(
    "deployments/free5gc/config/nrfcfg.yaml"
    "deployments/free5gc/config/amfcfg.yaml"
    "deployments/free5gc/config/smfcfg.yaml"
    "deployments/free5gc/config/upfcfg.yaml"
    "deployments/free5gc/config/udrcfg.yaml"
    "deployments/free5gc/config/udmcfg.yaml"
    "deployments/free5gc/config/ausfcfg.yaml"
    "deployments/free5gc/config/nssfcfg.yaml"
    "deployments/free5gc/config/pcfcfg.yaml"
)

missing_files=0
for file in "${config_files[@]}"; do
    if [[ -f "$file" ]]; then
        echo "  [OK] $file"
    else
        echo "  [MISSING] $file"
        missing_files=$((missing_files + 1))
    fi
done

# Check docker-compose configuration
echo "Checking docker-compose services..."
services=(
    "db"
    "free5gc-nrf"
    "free5gc-amf"
    "free5gc-smf"
    "free5gc-upf"
    "free5gc-udr"
    "free5gc-udm"
    "free5gc-ausf"
    "free5gc-nssf"
    "free5gc-pcf"
)

missing_services=0
for service in "${services[@]}"; do
    if docker-compose config | grep -q "$service:"; then
        echo "  [OK] $service service defined"
    else
        echo "  [MISSING] $service service"
        missing_services=$((missing_services + 1))
    fi
done

# Summary
echo ""
echo "=== Test Summary ==="
if [[ $missing_files -eq 0 && $missing_services -eq 0 ]]; then
    echo "  [SUCCESS] All free5GC configuration is ready!"
    echo "  To start: make free5gc-start"
    exit 0
else
    echo "  [ERROR] Found $missing_files missing files and $missing_services missing services"
    exit 1
fi
