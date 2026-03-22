#!/bin/bash
# Script para generar un archivo Docker Compose con cantidad configurable de clientes
# Uso: ./generar-compose.sh <archivo_salida> <cantidad_clientes>

set -e

OUTFILE="$1"
NUM_CLIENTS="$2"

if [[ -z "$OUTFILE" || -z "$NUM_CLIENTS" ]]; then
	echo "Uso: $0 docker-compose-dev.yaml <cantidad_clientes>"
	exit 1
fi


cat > "$OUTFILE" <<EOF
name: tp0
services:
    server:
        container_name: server
        image: server:latest
        entrypoint: python3 /main.py
        environment:
            - PYTHONUNBUFFERED=1
        networks:
            - testing_net
        volumes:
            - ./server/config.ini:/config.ini
EOF


# Generar clientes
for i in $(seq 1 "$NUM_CLIENTS"); do
	cat >> "$OUTFILE" <<EOF
    client$i:
        container_name: client$i
        image: client:latest
        entrypoint: /client
        environment:
            - CLI_ID=$i
            - NAME=name$i
            - SURNAME=surname$i
            - DNI=$((1000 + i))
            - BIRTHDAY=1990-01-01
            - NUMBER=$((100 + i))
        networks:
            - testing_net 
        depends_on:
            - server
        volumes:
            - ./client/config.yaml:/config.yaml
            - ./.data/agency-$i.csv:/data/agency-$i.csv
EOF
done

# Definir redes
cat >> "$OUTFILE" <<EOF
networks:
    testing_net:
        ipam:
            driver: default
            config:
               - subnet: 172.25.125.0/24
EOF

echo "compose en $OUTFILE generado con $NUM_CLIENTS clientes."
