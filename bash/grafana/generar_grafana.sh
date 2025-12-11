#!/bin/bash

# Ubicación donde está tu docker-compose.yml
DASHBOARD_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../dashboard" && pwd)"

echo "Iniciando Grafana usando docker-compose en:"
echo "  $DASHBOARD_DIR"
echo ""

# Ir al directorio del docker-compose
cd "$DASHBOARD_DIR" || {
    echo "Error: no se pudo acceder al directorio $DASHBOARD_DIR"
    exit 1
}

# Levantar grafana
docker compose pull
docker compose up -d

echo ""
echo "Grafana está ejecutándose en http://localhost:3000"
