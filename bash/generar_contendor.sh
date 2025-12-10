#!/bin/bash

IMAGES=("alto_cpu_img" "alta_ram_im" "bajo_consumo_img")

for i in {1..10}
do
    # Imagen aleatoria
    IMG=${IMAGES[$RANDOM % ${#IMAGES[@]}]}

    # Nombre aleatorio
    NAME="auto_container_$(date +%s)_$RANDOM"

    echo "Creando contenedor: $NAME con imagen: $IMG"

    docker run -d --name "$NAME" "$IMG"
done
