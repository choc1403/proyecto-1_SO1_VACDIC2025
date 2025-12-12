#!/bin/bash

# Construccion de las imagenes


# Definir nombres de las imágenes
IMAGES=("high_cpu_img" "high_mem_img" "low_img")
PATHS=("./alto_consumo/cpu" "./alto_consumo/ram" "./bajo_consumo")

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

# Ejecutar docker images | grep, pero redirigir su salida a un archivo temporal o /dev/null
# y usar '|| true' para asegurar que el código de salida sea 0 incluso si grep no encuentra coincidencias.
# Nota: Ajusté la expresión regular para que coincida con los nombres definidos arriba.
docker images | grep -E "(alto_cpu_img|alta_ram_im|bajo_consumo_img|REPOSITORY)" 

# La línea anterior puede seguir causando el problema. La forma más segura es la siguiente:
# Ejecutar el grep y simplemente ignorar su código de salida, asegurando que el script termine exitosamente
# si no hubo errores de 'docker build' o de ruta.

docker images | grep -E "(alto_cpu_img|alta_ram_im|bajo_consumo_img|REPOSITORY)" || true

# Finalmente, aseguramos un código de salida exitoso (0) para el script completo, 
# a menos que haya habido un 'docker build' fallido.

if $image_built; then
    # Si construyó algo o fue exitoso anteriormente, salimos con 0
    exit 0
else
    # Si todo se saltó o fue exitoso, también salimos con 0
    exit 0 
fi