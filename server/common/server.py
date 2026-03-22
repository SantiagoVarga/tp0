
import socket
import logging
import struct
import json
from .utils import Bet, store_bets


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        # TODO: Modify this program to handle signal to graceful shutdown
        # the server
        while self._running:
            client_sock = self.__accept_new_connection()
            addr = client_sock.getpeername()
            self.__handle_client_connection(client_sock)

    def close_resources(self):
        logging.info("Cerrando socket del servidor...")
        try:
            self._server_socket.close()
            logging.info("Socket del servidor cerrado correctamente.")
        except Exception as e:
            logging.error(f"Error al cerrar socket del servidor: {e}")
        self._running = False

    def __handle_client_connection(self, client_sock):
       """
        Recibe un mensaje (apuesta individual o batch), lo procesa y responde.
        Evita short-read/write usando prefijo de longitud.
        """
        try:
            raw_len = self._recv_all(client_sock, 4)
            if not raw_len:
                raise Exception("No se pudo leer la longitud del mensaje")
            msg_len = struct.unpack('>I', raw_len)[0]
            msg_bytes = self._recv_all(client_sock, msg_len)
            msg = msg_bytes.decode('utf-8')
            
            # Detectar si es un batch o una apuesta individual
            if msg.startswith("BET/BATCH/"):
                response, cantidad = self.process_batch(msg)
                logging.info(f"action: apuesta_recibida | result: {'success' if 'SUCCESS' in response else 'fail'} | cantidad: {cantidad}")
            else:
                response = self.process_bet(msg)
            
            response_bytes = response.encode('utf-8')
            client_sock.sendall(struct.pack('>I', len(response_bytes)) + response_bytes)
        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()
            
    def process_bet(self, message):
        """
        Procesa una apuesta individual.
        Formato: BET/agencia/nombre/apellido/documento/nacimiento/numero
        Devuelve: RESPONSE/SUCCESS/... o RESPONSE/FAIL/...
        """
        try:
            parts = message.split('/')
            if len(parts) != 7 or parts[0] != "BET":
                raise ValueError("Formato de apuesta inválido")
            
            bet = Bet(
                agency=parts[1],
                first_name=parts[2],
                last_name=parts[3],
                document=parts[4],
                birthdate=parts[5],
                number=parts[6]
            )
            store_bets([bet])
            logging.info(f"action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}")
            return "RESPONSE/SUCCESS/Apuesta almacenada"
        except Exception as e:
            logging.error(f"action: apuesta_almacenada | result: fail | error: {e}")
            return "RESPONSE/FAIL/Error al almacenar apuesta"

    def process_batch(self, message) -> tuple[str, int]:
        """
        Procesa un batch de apuestas.
        Formato: BET/BATCH/{cantidad}/{agencia1}/{nombre1}/{apellido1}/{dni1}/{nacimiento1}/{numero1}/{agencia2}/...
        Devuelve: (RESPONSE/SUCCESS/... o RESPONSE/FAIL/..., cantidad_procesada)
        """
        try:
            parts = message.split('/')
            
            # Validar prefijo y cantidad
            if len(parts) < 3 or parts[0] != "BET" or parts[1] != "BATCH":
                raise ValueError("Formato de batch inválido")
            
            try:
                batch_size = int(parts[2])
            except ValueError:
                raise ValueError("Cantidad de apuestas inválida")
            
            # Validar que haya exactamente batch_size apuestas (6 campos cada una)
            expected_parts = 3 + (batch_size * 6)
            if len(parts) != expected_parts:
                raise ValueError(f"Cantidad de apuestas no coincide: esperadas {batch_size}, obtenidas {(len(parts) - 3) // 6}")
            
            bets = []
            
            # Extraer y crear objetos Bet
            for i in range(batch_size):
                start_idx = 3 + (i * 6)
                bet = Bet(
                    agency=parts[start_idx],
                    first_name=parts[start_idx + 1],
                    last_name=parts[start_idx + 2],
                    document=parts[start_idx + 3],
                    birthdate=parts[start_idx + 4],
                    number=parts[start_idx + 5]
                )
                bets.append(bet)
            
            # Almacenar todas las apuestas (si falla una, falla todo)
            store_bets(bets)
            
            # Loguear cada apuesta almacenada
            for bet in bets:
                logging.info(f"action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}")
            
            return "RESPONSE/SUCCESS/Batch procesado", batch_size
        
        except Exception as e:
            logging.error(f"action: apuesta_almacenada | result: fail | error: {e}")
            # Extraer cantidad si es posible para el log
            try:
                batch_size = int(message.split('/')[2])
            except:
                batch_size = 0
            return f"RESPONSE/FAIL/{str(e)}", batch_size

    def _recv_all(self, sock, n):
        data = b''
        while len(data) < n:
            packet = sock.recv(n - len(data))
            if not packet:
                return None
            data += packet
        return data

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
