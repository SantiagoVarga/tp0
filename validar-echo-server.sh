#!/bin/bash


# Mensaje de prueba
TEST_MSG="echo_test_$(date +%s)"

# Nombre de la red definida en docker-compose 
NETWORK="tp0_testing_net"
# Nombre del servicio del servidor
SERVER="server"
# Puerto del servidor
PORT=12345

# Ejecutar netcat en un contenedor temporal en la red de docker-compose
RESPONSE=$(docker run --rm --network "$NETWORK" alpine sh -c "echo $TEST_MSG | nc $SERVER $PORT")

if [ "$RESPONSE" = "$TEST_MSG" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi
