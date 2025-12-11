#!/bin/bash

# Asegurar que el script se ejecute como root (opcional, si solo se llama con sudo)
# if [ "$EUID" -ne 0 ]; then
#   echo "Por favor, ejecute como root (con sudo)"
#   exit 1
# fi

CRONFILE="/etc/cron.d/project_containers_so1"
SCRIPT_PATH="/home/juan/Escritorio/proyecto-1_SO1_VACDIC2025/bash/generar_contenedor.sh"
LOG_PATH="/var/log/project_containers.log"



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
