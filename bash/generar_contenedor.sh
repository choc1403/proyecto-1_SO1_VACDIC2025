#!/bin/bash

# Definir las imágenes disponibles
LOW_IMAGE="low_img"
HIGH_CPU_IMAGE="high_cpu_img"
HIGH_RAM_IMAGE="high_mem_img"

# Crear un array con las imágenes
IMAGES=("$LOW_IMAGE" "$HIGH_CPU_IMAGE" "$HIGH_RAM_IMAGE")

# Número de contenedores a crear
NUM_CONTAINERS=10

# Nombres base para los contenedores
CONTAINER_BASE_NAME="random_container"

echo "Creando $NUM_CONTAINERS contenedores Docker aleatorios..."
echo "========================================================"

for ((i=1; i<=$NUM_CONTAINERS; i++))
do
    # Seleccionar imagen aleatoria
    RANDOM_INDEX=$((RANDOM % ${#IMAGES[@]}))
    SELECTED_IMAGE=${IMAGES[$RANDOM_INDEX]}
    
    # Generar nombre único para el contenedor
    CONTAINER_NAME="${CONTAINER_BASE_NAME}_${i}_$(date +%s%N | md5sum | head -c 6)"
    
    echo "Creando contenedor $i: $CONTAINER_NAME con imagen: $SELECTED_IMAGE"
    
    # Comando para crear el contenedor Docker (ejecuta en segundo plano)
    docker run -d --name "$CONTAINER_NAME" "$SELECTED_IMAGE"
   
    
    echo "  ✅ Contenedor creado: $CONTAINER_NAME"
    echo "----------------------------------------"
done

echo ""
echo "Resumen:"
echo "========="
echo "Contenedores creados: $NUM_CONTAINERS"
echo "Imágenes utilizadas:"
echo "  - $LOW_IMAGE"
echo "  - $HIGH_CPU_IMAGE"
echo "  - $HIGH_RAM_IMAGE"
