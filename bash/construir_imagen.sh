#!/bin/bash

# Construccion de las imagenes


# Definir nombres de las imágenes
IMAGES=("alto_cpu_img" "alta_ram_im" "bajo_consumo_img")
PATHS=("../bash/alto_consumo/cpu" "../bash/alto_consumo/ram" "../bash/bajo_consumo")

echo "Verificando imágenes Docker..."

# Bandera para rastrear si alguna imagen fue construida
image_built=false

# Recorrer todas las imágenes
for i in "${!IMAGES[@]}"; do
    IMAGE_NAME="${IMAGES[$i]}"
    IMAGE_PATH="${PATHS[$i]}"
    
    # Verificar si la imagen ya existe
    if docker image inspect "$IMAGE_NAME" > /dev/null 2>&1; then
        echo "La imagen '$IMAGE_NAME' ya existe. Saltando construcción."
    else
        echo "Construyendo imagen '$IMAGE_NAME' desde '$IMAGE_PATH'..."
        
        # Verificar si el directorio existe
        if [ ! -d "$IMAGE_PATH" ]; then
            echo "Error: El directorio '$IMAGE_PATH' no existe."
            continue
        fi
        
        # Construir la imagen
        docker build -t "$IMAGE_NAME" "$IMAGE_PATH"
        
        if [ $? -eq 0 ]; then
            echo "Imagen '$IMAGE_NAME' construida exitosamente."
            image_built=true
        else
            echo "Error al construir la imagen '$IMAGE_NAME'."
        fi
    fi
done

echo ""
echo "Resumen:"
echo "Todas las imágenes han sido verificadas."

# Mostrar imágenes creadas
echo ""
echo "Imágenes Docker existentes con los nombres especificados:"
docker images | grep -E "(high_cpu_image|high_ram_image|low_usage_image|REPOSITORY)"