
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
        Recibe una apuesta serializada, la procesa y responde al cliente.
        Evita short-read/write usando prefijo de longitud.
        """
        try:
            raw_len = self._recv_all(client_sock, 4)
            if not raw_len:
                raise Exception("No se pudo leer la longitud del mensaje")
            msg_len = struct.unpack('>I', raw_len)[0]
            msg_bytes = self._recv_all(client_sock, msg_len)
            msg = msg_bytes.decode('utf-8')
            status, dni, numero = self.process_bet(msg)
            if status == "success":
                response = f"RESPONSE/SUCCESS/Apuesta almacenada"
            else:
                response = f"RESPONSE/FAIL/Error al almacenar apuesta"
            response_bytes = response.encode('utf-8')
            client_sock.sendall(struct.pack('>I', len(response_bytes)) + response_bytes)
        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()

    def process_bet(self, message):
        """
        Procesa el mensaje de apuesta, crea el objeto Bet y almacena la apuesta.
        El mensaje debe tener los campos separados por '/': agencia/nombre/apellido/documento/nacimiento/numero
        Devuelve status, dni y numero para logging y respuesta.
        """
        try:
            parts = message.split('/')
            if len(parts) != 6:
                raise ValueError("Formato de apuesta inválido")
            bet = Bet(
                agency=parts[0],
                first_name=parts[1],
                last_name=parts[2],
                document=parts[3],
                birthdate=parts[4],
                number=parts[5]
            )
            store_bets([bet])
            logging.info(f"action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}")
            return "success", bet.document, bet.number
        except Exception as e:
            logging.error(f"action: apuesta_almacenada | result: fail | error: {e}")
            return "fail", None, None

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
