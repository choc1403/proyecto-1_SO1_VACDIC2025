#!/bin/bash

# Definir las imágenes
LOW_IMAGE="low_img"
HIGH_CPU_IMAGE="high_cpu_img"
HIGH_RAM_IMAGE="high_mem_img"

# Crear un array con las imágenes
IMAGES=("$LOW_IMAGE" "$HIGH_CPU_IMAGE" "$HIGH_RAM_IMAGE")

# Número máximo de contenedores totales
MAX_CONTAINERS=10

# Función para contar contenedores de las imágenes especificadas
contar_contenedores() {
    # Contar contenedores de cada imagen (incluyendo los detenidos)
    local total=0
    
    for imagen in "${IMAGES[@]}"; do
        # Contar contenedores creados de cada imagen
        local cantidad=$(docker ps -a --filter "ancestor=$imagen" --format "{{.ID}}" | wc -l)
        total=$((total + cantidad))
    done
    
    echo $total
}

# Verificar si Docker está instalado y funcionando
if ! command -v docker &> /dev/null; then
    echo "Error: Docker no está instalado o no se encuentra en el PATH."
    exit 1
fi

# Verificar si el daemon de Docker está ejecutándose
if ! docker info &> /dev/null; then
    echo "Error: El daemon de Docker no está ejecutándose."
    exit 1
fi

# Obtener el número actual de contenedores de las imágenes especificadas
contenedores_actuales=$(contar_contenedores)

echo "Contenedores actuales de las imágenes especificadas: $contenedores_actuales"
echo "Límite máximo de contenedores: $MAX_CONTAINERS"

# Verificar si ya se alcanzó el límite de contenedores
if [ "$contenedores_actuales" -ge "$MAX_CONTAINERS" ]; then
    echo "Ya se alcanzó el límite de $MAX_CONTAINERS contenedores."
    echo "No se creará un nuevo contenedor."
    
    # Mostrar información de los contenedores existentes
    echo -e "\nContenedores existentes:"
    for imagen in "${IMAGES[@]}"; do
        echo "Imagen: $imagen"
        docker ps -a --filter "ancestor=$imagen" --format "table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.CreatedAt}}"
        echo ""
    done
    
    exit 0
fi

# Si no se ha alcanzado el límite, crear un nuevo contenedor

# Seleccionar una imagen aleatoria del array
indice_imagen=$((RANDOM % ${#IMAGES[@]}))
IMAGEN_SELECCIONADA=${IMAGES[$indice_imagen]}

# Generar un nombre único para el contenedor
NOMBRE_CONTENEDOR="${IMAGEN_SELECCIONADA}_$(date +%Y%m%d_%H%M%S)_$RANDOM"

echo "Creando un nuevo contenedor..."
echo "Imagen seleccionada: $IMAGEN_SELECCIONADA"
echo "Nombre del contenedor: $NOMBRE_CONTENEDOR"

# Crear el contenedor (ajusta los parámetros según tus necesidades)
# Nota: Asegúrate de que las imágenes existan localmente o en un registro
case $IMAGEN_SELECCIONADA in
    "$LOW_IMAGE")
        # Configuración para contenedor de baja demanda
        docker run -d --name "$NOMBRE_CONTENEDOR" "$IMAGEN_SELECCIONADA"
        ;;
    "$HIGH_CPU_IMAGE")
        # Configuración para contenedor de alto CPU
        docker run -d --name "$NOMBRE_CONTENEDOR"  "$IMAGEN_SELECCIONADA"
        ;;
    "$HIGH_RAM_IMAGE")
        # Configuración para contenedor de alta memoria
        docker run -d --name "$NOMBRE_CONTENEDOR"  "$IMAGEN_SELECCIONADA"
        ;;
    *)
        # Configuración por defecto
        docker run -d --name "$NOMBRE_CONTENEDOR" "$IMAGEN_SELECCIONADA"
        ;;
esac

# Verificar si el contenedor se creó correctamente
if [ $? -eq 0 ]; then
    echo "Contenedor creado exitosamente: $NOMBRE_CONTENEDOR"
    
    # Mostrar el nuevo total de contenedores
    nuevo_total=$(contar_contenedores)
    echo "Total de contenedores ahora: $nuevo_total"
else
    echo "Error al crear el contenedor."
    exit 1
fi

# Mostrar información del contenedor recién creado
echo -e "\nInformación del contenedor creado:"
docker ps --filter "name=$NOMBRE_CONTENEDOR" --format "table {{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"