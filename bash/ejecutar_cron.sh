#!/bin/bash

# --- VARIABLES DE CONFIGURACIÓN ---
CRONFILE="/etc/cron.d/project_containers_so1"
SCRIPT_PATH="/home/juan/Escritorio/proyecto-1_SO1_VACDIC2025/bash/generar_contenedor.sh"
# Ya tienes la ruta al script de construcción de imagen:
SCRIPT_PATH_IMG="/home/juan/Escritorio/proyecto-1_SO1_VACDIC2025/bash/construir_imagen.sh"
LOG_PATH="/var/log/project_containers.log"
# -----------------------------------


# 1. CONSTRUIR LAS IMÁGENES (Ejecución Única)
echo "Ejecutando script de construcción de imágenes: $SCRIPT_PATH_IMG"
# Se ejecuta el script y se redirige la salida a STDOUT
bash "$SCRIPT_PATH_IMG"

if [ $? -eq 0 ]; then
    echo "Imágenes Docker construidas exitosamente."
else
    echo "Error al construir las imágenes Docker. Revise el script y los logs."
    exit 1 # Detener el proceso si la construcción falla
fi


# 2. CONFIGURAR EL CRONJOB (Para generar/iniciar contenedores cada minuto)
echo "Configurando cronjob para generar contenedores cada minuto..."
# El cronjob solo necesita el script de generación de contenedores
echo "* * * * * root $SCRIPT_PATH >$LOG_PATH 2>&1" | sudo tee $CRONFILE > /dev/null

# Cambiar permisos (requiere sudo)
sudo chmod 0644 $CRONFILE

# Recargar el servicio de cron (requiere sudo)
if command -v systemctl > /dev/null; then
  sudo systemctl restart cron
elif command -v service > /dev/null; then
 sudo service cron reload
else
 echo "Advertencia: No se pudo reiniciar el servicio cron."
fi

echo "Cronjob instalado y servicio reiniciado."