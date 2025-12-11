#!/bin/bash

DASHBOARD_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../dashboard" && pwd)"

echo "Deteniendo Grafana."
echo ""

cd "$DASHBOARD_DIR" || {
    echo "Error: no se pudo acceder al directorio $DASHBOARD_DIR"
    exit 1
}

docker compose down

echo ""
echo "Grafana detenido."
