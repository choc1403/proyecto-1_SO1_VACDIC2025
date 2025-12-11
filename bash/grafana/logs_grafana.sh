#!/bin/bash

echo "Mostrando los últimos 200 logs de Grafana:"
echo ""
docker logs --tail 200 grafana_so1
