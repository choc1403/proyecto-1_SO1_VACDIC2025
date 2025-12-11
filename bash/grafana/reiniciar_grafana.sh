#!/bin/bash

DASHBOARD_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../dashboard" && pwd)"

echo "Reiniciando Grafana."
echo ""

cd "$DASHBOARD_DIR" || {
    echo "Error: no se pudo acceder al directorio $DASHBOARD_DIR"
    exit 1
}

docker compose down
docker compose up -d --build

echo ""
echo "Grafana reiniciado. Disponible en http://localhost:3000"
