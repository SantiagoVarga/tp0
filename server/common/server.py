import logging
import socket
import struct

from .utils import Bet, store_bets


class Server:
    def __init__(self, port, listen_backlog):
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self._server_socket.bind(("", port))
        self._server_socket.listen(listen_backlog)

    def run(self):
        while True:
            client_sock = self.__accept_new_connection()
            self.__handle_client_connection(client_sock)

    def close_resources(self):
        try:
            self._server_socket.close()
        except Exception:
            pass

    def __handle_client_connection(self, client_sock):
        try:
            raw_len = self._recv_all(client_sock, 4)
            if not raw_len:
                raise RuntimeError("client disconnected before length")

            msg_len = struct.unpack(">I", raw_len)[0]
            msg = self._recv_all(client_sock, msg_len).decode("utf-8")

            if msg.startswith("BET/BATCH/"):
                response, cantidad, ok = self.process_batch(msg)
                logging.info(
                    f"action: apuesta_recibida | result: {'success' if ok else 'fail'} | cantidad: {cantidad}"
                )
            else:
                # individual bet (optional for ej6, but keeps compatibility)
                response, _dni, _num, ok = self.process_bet(msg)

            response_bytes = response.encode("utf-8")
            client_sock.sendall(struct.pack(">I", len(response_bytes)) + response_bytes)
        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            try:
                client_sock.close()
            except Exception:
                pass

    def process_bet(self, message: str):
        """
        Formato: BET/agencia/nombre/apellido/dni/nacimiento/numero
        """
        try:
            parts = message.split("/")
            if len(parts) != 7 or parts[0] != "BET":
                raise ValueError("invalid bet format")

            bet = Bet(
                agency=parts[1],
                first_name=parts[2],
                last_name=parts[3],
                document=parts[4],
                birthdate=parts[5],
                number=parts[6],
            )

            store_bets([bet])
            logging.info(
                f"action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}"
            )
            return "RESPONSE/SUCCESS/Apuesta almacenada", bet.document, bet.number, True
        except Exception as e:
            logging.error(f"action: apuesta_almacenada | result: fail | error: {e}")
            return "RESPONSE/FAIL/Error al almacenar apuesta", "", "", False

    def process_batch(self, message: str):
        """
        Formato: BET/BATCH/{cantidad}/{agency}/{name}/{surname}/{dni}/{birth}/{number}...
        """
        try:
            parts = message.split("/")
            if len(parts) < 3 or parts[0] != "BET" or parts[1] != "BATCH":
                raise ValueError("invalid batch format")

            cantidad = int(parts[2])
            expected = 3 + (cantidad * 6)
            if len(parts) != expected:
                raise ValueError("batch size mismatch")

            bets = []
            for i in range(cantidad):
                base = 3 + i * 6
                bet = Bet(
                    agency=parts[base],
                    first_name=parts[base + 1],
                    last_name=parts[base + 2],
                    document=parts[base + 3],
                    birthdate=parts[base + 4],
                    number=parts[base + 5],
                )
                bets.append(bet)

            # Éxito sólo si store_bets no falla para TODO el batch
            store_bets(bets)

            # logs por apuesta almacenada (si tu enunciado/TP previo lo sigue pidiendo)
            for bet in bets:
                logging.info(
                    f"action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}"
                )

            return "RESPONSE/SUCCESS/Batch procesado", cantidad, True
        except Exception as e:
            # cantidad para log aun si el batch está mal formado
            try:
                cantidad = int(message.split("/")[2])
            except Exception:
                cantidad = 0
            logging.error(f"action: apuesta_almacenada | result: fail | error: {e}")
            return "RESPONSE/FAIL/Batch invalido", cantidad, False

    def _recv_all(self, sock, n: int) -> bytes:
        data = b""
        while len(data) < n:
            chunk = sock.recv(n - len(data))
            if not chunk:
                return b""
            data += chunk
        return data

    def __accept_new_connection(self):
        logging.info("action: accept_connections | result: in_progress")
        c, addr = self._server_socket.accept()
        logging.info(f"action: accept_connections | result: success | ip: {addr[0]}")
        return c