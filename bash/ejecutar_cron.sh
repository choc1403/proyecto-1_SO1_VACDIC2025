#!/bin/bash
CRONFILE="/etc/cron.d/project_containers_so1"
# The cronjob runs as root — change if you want a different user
echo "* * * * * root /home/juan/Escritorio/proyecto-1_SO1_VACDIC2025/bash/generar_contenedor.sh >/var/log/project_containers.log 2>&1" > $CRONFILE
chmod 644 $CRONFILE
service cron reload || systemctl restart cron || true
echo "cron started"
