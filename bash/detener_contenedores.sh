#!/bin/bash

# Detener todos los contenedores en ejecución
echo "Deteniendo todos los contenedores..."
running_containers=$(docker ps -q)

if [ -n "$running_containers" ]; then
    docker stop $running_containers
    echo "Contenedores detenidos."
else
    echo "No hay contenedores en ejecución."
fi

# Eliminar todos los contenedores
echo "Eliminando todos los contenedores..."
all_containers=$(docker ps -aq)

if [ -n "$all_containers" ]; then
    docker rm $all_containers
    echo "Contenedores eliminados."
else
    echo "No hay contenedores para eliminar."
fi

echo "Tarea completada."
