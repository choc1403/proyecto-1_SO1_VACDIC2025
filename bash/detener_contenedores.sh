#!/bin/bash

EXCLUDE_NAME="grafana_so1"

echo "Buscando contenedores a detener y eliminar (excepto $EXCLUDE_NAME)..."

# Obtener contenedores en ejecución EXCEPTO grafana_so1
running_containers=$(docker ps -q --filter "name!=${EXCLUDE_NAME}")

echo "Deteniendo contenedores..."
if [ -n "$running_containers" ]; then
    docker stop $running_containers
    echo "Contenedores detenidos."
else
    echo "No hay contenedores para detener (o solo está $EXCLUDE_NAME)."
fi

# Obtener todos los contenedores EXCEPTO grafana_so1
all_containers=$(docker ps -aq --filter "name!=${EXCLUDE_NAME}")

echo "Eliminando contenedores..."
if [ -n "$all_containers" ]; then
    docker rm $all_containers
    echo "Contenedores eliminados."
else
    echo "No hay contenedores para eliminar (o solo está $EXCLUDE_NAME)."
fi

echo "Tarea completada. El contenedor '$EXCLUDE_NAME' sigue intacto."
