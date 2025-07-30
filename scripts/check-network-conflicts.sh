#!/bin/bash

# Script to check for Docker network conflicts before deployment

set -e

echo "🔍 Checking for Docker network conflicts..."

# Get the subnet from docker-compose file
COMPOSE_SUBNET=$(grep -A 3 "ipam:" docker-compose.prod.yml | grep "subnet:" | awk '{print $3}' | head -1)

if [ -z "$COMPOSE_SUBNET" ]; then
    echo "❌ Could not find subnet configuration in docker-compose.prod.yml"
    exit 1
fi

echo "📋 Target subnet: $COMPOSE_SUBNET"

# Check existing networks
echo ""
echo "🌐 Existing Docker networks:"
docker network ls --format "table {{.Name}}\t{{.Driver}}\t{{.Scope}}"

echo ""
echo "📊 Network IP ranges:"
docker network inspect $(docker network ls -q) --format='{{.Name}}: {{range .IPAM.Config}}{{.Subnet}}{{end}}' 2>/dev/null | grep -v "^.*: $" | sort

# Check for conflicts
echo ""
echo "🔍 Checking for subnet conflicts..."

CONFLICT_FOUND=false
while IFS= read -r line; do
    network_name=$(echo "$line" | cut -d: -f1)
    network_subnet=$(echo "$line" | cut -d: -f2 | tr -d ' ')
    
    if [ "$network_subnet" = "$COMPOSE_SUBNET" ] && [ "$network_name" != "qr-code-digital-signature_app-network" ]; then
        echo "⚠️  CONFLICT: Network '$network_name' is using subnet $network_subnet"
        CONFLICT_FOUND=true
    fi
done < <(docker network inspect $(docker network ls -q) --format='{{.Name}}: {{range .IPAM.Config}}{{.Subnet}}{{end}}' 2>/dev/null | grep -v "^.*: $")

if [ "$CONFLICT_FOUND" = true ]; then
    echo ""
    echo "❌ Network conflicts detected!"
    echo "💡 Suggested solutions:"
    echo "   1. Change the subnet in docker-compose.prod.yml"
    echo "   2. Remove conflicting networks: docker network rm <network_name>"
    echo "   3. Run the deployment script which includes automatic cleanup"
    exit 1
else
    echo "✅ No network conflicts detected!"
    echo "🚀 Safe to proceed with deployment"
fi

echo ""
echo "📈 Available subnet ranges (not in use):"
USED_RANGES=$(docker network inspect $(docker network ls -q) --format='{{range .IPAM.Config}}{{.Subnet}}{{end}}' 2>/dev/null | grep -v "^$" | sort -u)

for i in {17..30}; do
    RANGE="172.$i.0.0/16"
    if ! echo "$USED_RANGES" | grep -q "$RANGE"; then
        echo "   - $RANGE"
    fi
done