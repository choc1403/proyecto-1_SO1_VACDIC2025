#!/bin/bash
CRONFILE="/etc/cron.d/project_containers_so1"
if [ -f "$CRONFILE" ]; then
    sudo rm -f "$CRONFILE"
    #service cron reload || systemctl restart cron || true
    echo "cron removed"
else
    echo "cron not found"
fi
