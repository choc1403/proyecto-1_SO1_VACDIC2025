#!/bin/bash

echo "Estado del contenedor Grafana:"
docker ps --filter "name=grafana_so1"
echo ""

if docker ps --filter "name=grafana_so1" --format "{{.Names}}" | grep -q "grafana_so1"; then
    echo "Grafana está corriendo."
else
    echo "Grafana NO está corriendo."
fi
