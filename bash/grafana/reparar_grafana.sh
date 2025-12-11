#!/bin/bash

DASHBOARD_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../dashboard" && pwd)"

echo "Reparando Grafana."

cd "$DASHBOARD_DIR" || { echo "Error accediendo a $DASHBOARD_DIR"; exit 1; }

docker compose down
docker system prune -f
docker compose up -d --build

echo ""
echo "Reparación completa. Grafana está en http://localhost:3000"
