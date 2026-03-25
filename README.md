# TP0: Docker + Comunicaciones + Concurrencia

En el presente repositorio se provee un esqueleto básico de cliente/servidor, en donde todas las dependencias del mismo se encuentran encapsuladas en containers. Los alumnos deberán resolver una guía de ejercicios incrementales, teniendo en cuenta las condiciones de entrega descritas al final de este enunciado.

 El cliente (Golang) y el servidor (Python) fueron desarrollados en diferentes lenguajes simplemente para mostrar cómo dos lenguajes de programación pueden convivir en el mismo proyecto con la ayuda de containers, en este caso utilizando [Docker Compose](https://docs.docker.com/compose/).

## Instrucciones de uso
El repositorio cuenta con un **Makefile** que incluye distintos comandos en forma de targets. Los targets se ejecutan mediante la invocación de:  **make \<target\>**. Los target imprescindibles para iniciar y detener el sistema son **docker-compose-up** y **docker-compose-down**, siendo los restantes targets de utilidad para el proceso de depuración.

Los targets disponibles son:

| target  | accion  |
|---|---|
|  `docker-compose-up`  | Inicializa el ambiente de desarrollo. Construye las imágenes del cliente y el servidor, inicializa los recursos a utilizar (volúmenes, redes, etc) e inicia los propios containers. |
| `docker-compose-down`  | Ejecuta `docker-compose stop` para detener los containers asociados al compose y luego  `docker-compose down` para destruir todos los recursos asociados al proyecto que fueron inicializados. Se recomienda ejecutar este comando al finalizar cada ejecución para evitar que el disco de la máquina host se llene de versiones de desarrollo y recursos sin liberar. |
|  `docker-compose-logs` | Permite ver los logs actuales del proyecto. Acompañar con `grep` para lograr ver mensajes de una aplicación específica dentro del compose. |
| `docker-image`  | Construye las imágenes a ser utilizadas tanto en el servidor como en el cliente. Este target es utilizado por **docker-compose-up**, por lo cual se lo puede utilizar para probar nuevos cambios en las imágenes antes de arrancar el proyecto. |
| `build` | Compila la aplicación cliente para ejecución en el _host_ en lugar de en Docker. De este modo la compilación es mucho más veloz, pero requiere contar con todo el entorno de Golang y Python instalados en la máquina _host_. |

### Servidor

Se trata de un "echo server", en donde los mensajes recibidos por el cliente se responden inmediatamente y sin alterar. 

Se ejecutan en bucle las siguientes etapas:

1. Servidor acepta una nueva conexión.
2. Servidor recibe mensaje del cliente y procede a responder el mismo.
3. Servidor desconecta al cliente.
4. Servidor retorna al paso 1.


### Cliente
 se conecta reiteradas veces al servidor y envía mensajes de la siguiente forma:
 
1. Cliente se conecta al servidor.
2. Cliente genera mensaje incremental.
3. Cliente envía mensaje al servidor y espera mensaje de respuesta.
4. Servidor responde al mensaje.
5. Servidor desconecta al cliente.
6. Cliente verifica si aún debe enviar un mensaje y si es así, vuelve al paso 2.

### Ejemplo

Al ejecutar el comando `make docker-compose-up`  y luego  `make docker-compose-logs`, se observan los siguientes logs:

```
client1  | 2024-08-21 22:11:15 INFO     action: config | result: success | client_id: 1 | server_address: server:12345 | loop_amount: 5 | loop_period: 5s | log_level: DEBUG
client1  | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:14 DEBUG    action: config | result: success | port: 12345 | listen_backlog: 5 | logging_level: DEBUG
server   | 2024-08-21 22:11:14 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°3
client1  | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°3
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°5
client1  | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°5
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:40 INFO     action: loop_finished | result: success | client_id: 1
client1 exited with code 0
```

### Ejercicio N°7:

Modificar los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas y así proceder con el sorteo.
Inmediatamente después de la notificacion, los clientes consultarán la lista de ganadores del sorteo correspondientes a su agencia.
Una vez el cliente obtenga los resultados, deberá imprimir por log: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`.

El servidor deberá esperar la notificación de las 5 agencias para considerar que se realizó el sorteo e imprimir por log: `action: sorteo | result: success`.
Luego de este evento, podrá verificar cada apuesta con las funciones `load_bets(...)` y `has_won(...)` y retornar los DNI de los ganadores de la agencia en cuestión. Antes del sorteo no se podrán responder consultas por la lista de ganadores con información parcial.

Las funciones `load_bets(...)` y `has_won(...)` son provistas por la cátedra y no podrán ser modificadas por el alumno.

No es correcto realizar un broadcast de todos los ganadores hacia todas las agencias, se espera que se informen los DNIs ganadores que correspondan a cada una de ellas.

### Resolución - Ejercicio N°7

Para cumplir con el enunciado del Ejercicio 7, se realizaron los siguientes cambios en esta branch:

---

#### 1. Cliente 

  - Al terminar de enviar todas las apuestas, el cliente envía un mensaje especial de notificación al servidor indicando que finalizó (`DONE/{agencia_id}`).
  - **Dónde:**  
    - Implementado en el cliente,`StartClientLoop` en `client/common/client.go`.
    - Serialización y envío del mensaje DONE en el módulo de comunicación (`client/common/comm.go`).
  - Inmediatamente después de notificar la finalización, el cliente consulta al servidor por la lista de ganadores de su agencia (`WINNERS/{agencia_id}`).
  - El cliente espera hasta recibir la respuesta del sorteo y, al obtenerla, imprime por log:
    ```
    action: consulta_ganadores | result: success | cant_ganadores: ${CANT}
    ```
  - **Dónde:**  
    - Lógica de consulta y log en el cliente, luego del envío de DONE (en `client/common/client.go`).
    - Comunicación y parsing de la respuesta en el módulo de comunicación (`client/common/comm.go`).

  ![Extracto: Client-comm.go](img/ej7-commGO.png)
  ![Extracto: Client-StartClientLoop](img/ej7-ClientLoop.png)
---


#### 2. Coordinación y sorteo en el servidor

- **Servidor:**  
  - El servidor espera recibir la notificación de finalización de las 5 agencias (o la cantidad configurada).
  - Una vez recibidas todas, realiza el sorteo usando las funciones provistas `load_bets(...)` y `has_won(...)`.
  - Al realizar el sorteo, imprime por log:
    ```
    action: sorteo | result: success
    ```
  - El servidor responde a las consultas de ganadores (`WINNERS/{agencia_id}`) **solo después** de haber realizado el sorteo. Antes de eso, responde con un mensaje indicando que el sorteo no está listo.
  - **Dónde:**  
    - Manejo de mensajes DONE y coordinación del sorteo en el método correspondiente del servidor (por ejemplo, `process_done` y `process_winners` en `server/common/server.py`).
    - Uso de las funciones `load_bets` y `has_won` (`server/common/utils.py`) en el sorteo.
  ![Extracto: Handle](img/ej7-Handle.png)
  ![Extracto: processDone](img/ej7-processDone.png)
  ![Extracto: processWinners](img/ej7-processWinners.png)

---

#### 3. Protocolo y respuestas

- Se agregaron nuevos tipos de mensajes al protocolo:
  - `DONE/{agencia_id}`: notificación de fin de apuestas.
  - `WINNERS/{agencia_id}`: consulta de ganadores.
  - Respuestas específicas para indicar éxito, sorteo no listo, y la lista de ganadores.
- **Dónde:**  
  - Definición y parsing de estos mensajes en los módulos de comunicación tanto del cliente como del servidor.

---

#### 4. Restricción de información

- El servidor responde a cada agencia **solo con los DNIs ganadores correspondientes a esa agencia**.
- No se realiza un broadcast global de ganadores.


---

**En resumen:**  
Se agregaron mensajes y lógica para notificar el fin de apuestas, coordinar el sorteo en el servidor, consultar ganadores por agencia y garantizar que cada cliente reciba solo la información correspondiente, cumpliendo con el protocolo y los logs requeridos por el enunciado.




