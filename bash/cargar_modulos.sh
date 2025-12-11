#!/bin/bash
# run as root
MODDIR="/home/juan/Escritorio/proyecto-1_SO1_VACDIC2025/modulo-kernel"
if [ -f "$MODDIR/sysinfo.ko" ]; then
    /sbin/insmod "$MODDIR/sysinfo.ko" || true
fi
if [ -f "$MODDIR/continfo.ko" ]; then
    /sbin/insmod "$MODDIR/continfo.ko" || true
fi
