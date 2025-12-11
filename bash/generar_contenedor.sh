#!/bin/bash

# Configuración
LOW_IMAGE="bajo_consumo_img"
HIGH_CPU_IMAGE="alto_cpu_img"
HIGH_RAM_IMAGE="alta_ram_im"

REQUIRED_LOW=3
REQUIRED_HIGH=2
REQUIRED_TOTAL=10

# 1. CONTAR EXISTENTES
CURRENT_LOW=$(docker ps -a --filter "ancestor=$LOW_IMAGE" --format "{{.ID}}" | wc -l)
CURRENT_HIGH_CPU=$(docker ps -a --filter "ancestor=$HIGH_CPU_IMAGE" --format "{{.ID}}" | wc -l)
CURRENT_HIGH_RAM=$(docker ps -a --filter "ancestor=$HIGH_RAM_IMAGE" --format "{{.ID}}" | wc -l)

CURRENT_HIGH=$((CURRENT_HIGH_CPU + CURRENT_HIGH_RAM))
CURRENT_TOTAL=$(docker ps -a --format "{{.ID}}" | wc -l)

echo "Actual:"
echo "   Bajo consumo:  $CURRENT_LOW"
echo "   Alto consumo:  $CURRENT_HIGH"
echo "   Total:         $CURRENT_TOTAL"

# ---------- 2. CREAR CONTENEDORES DE BAJO CONSUMO ----------
MISSING_LOW=$((REQUIRED_LOW - CURRENT_LOW))

if (( MISSING_LOW > 0 )); then
    echo "Creando $MISSING_LOW contenedores de BAJO consumo..."
    for i in $(seq 1 $MISSING_LOW)
    do
        NAME="low_container_$(date +%s)_$RANDOM"
        docker run -d --name "$NAME" "$LOW_IMAGE"
    done
fi

# ---------- 3. CREAR CONTENEDORES DE ALTO CONSUMO ----------
MISSING_HIGH=$((REQUIRED_HIGH - CURRENT_HIGH))

if (( MISSING_HIGH > 0 )); then
    echo "Creando $MISSING_HIGH contenedores de ALTO consumo..."
    for i in $(seq 1 $MISSING_HIGH)
    do
        NAME="high_container_$(date +%s)_$RANDOM"

        # Elegir aleatorio entre alto CPU y alto RAM
        if (( RANDOM % 2 == 0 )); then
            IMG="$HIGH_CPU_IMAGE"
        else
            IMG="$HIGH_RAM_IMAGE"
        fi

        docker run -d --name "$NAME" "$IMG"
    done
fi

# ---------- 4. COMPLETAR HASTA 10 CONTENEDORES TOTALES ----------
CURRENT_TOTAL=$(docker ps -a --format "{{.ID}}" | wc -l)
MISSING_TOTAL=$((REQUIRED_TOTAL - CURRENT_TOTAL))

if (( MISSING_TOTAL > 0 )); then
    echo "Creando $MISSING_TOTAL contenedores para completar 10..."

    ALL_IMAGES=("$LOW_IMAGE" "$HIGH_CPU_IMAGE" "$HIGH_RAM_IMAGE")

    for i in $(seq 1 $MISSING_TOTAL)
    do
        IMG=${ALL_IMAGES[$RANDOM % ${#ALL_IMAGES[@]}]}
        NAME="auto_container_$(date +%s)_$RANDOM"
        docker run -d --name "$NAME" "$IMG"
    done
fi

echo "Listo. Sistema estable según las reglas."
