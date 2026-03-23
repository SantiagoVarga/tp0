import logging
import socket
import struct
import threading
from concurrent.futures import ThreadPoolExecutor

from .utils import Bet, store_bets, load_bets, has_won
import os


class Server:
    def __init__(self, port, listen_backlog):
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self._server_socket.bind(("", port))
        self._server_socket.listen(listen_backlog)

        self._required_agencies = int(os.getenv("AGENCIES_COUNT", "0"))
        if self._required_agencies <= 0:
            logging.warning("action: config | result: fail | error: AGENCIES_COUNT missing/invalid")

       
        self._done_agencies = set()
        self._draw_done = False
        self._drawing_in_progress = False
        self._winners_by_agency = {}

        # Thread pool (evita 1 thread ilimitado por cliente)
        workers = int(os.getenv("SERVER_WORKERS", "32"))
        self._executor = ThreadPoolExecutor(max_workers=workers)



    def run(self):
        while True:
            client_sock = self.__accept_new_connection()
            self._executor.submit(self.__handle_client_connection, client_sock)
    def close_resources(self):
        try:
            self._executor.shutdown(wait=False, cancel_futures=True)
        except Exception:
            pass
        try:
            self._server_socket.close()
        except Exception:
            pass

    def __handle_client_connection(self, client_sock):
        try:
            while True:
                raw_len = self._recv_all(client_sock, 4)
                if not raw_len:
                    return

                msg_len = struct.unpack(">I", raw_len)[0]
                msg = self._recv_all(client_sock, msg_len).decode("utf-8")

                if msg.startswith("BET/BATCH/"):
                    response, cantidad, ok = self.process_batch(msg)
                    logging.info(
                        f"action: apuesta_recibida | result: {'success' if ok else 'fail'} | cantidad: {cantidad}"
                    )
                elif msg.startswith("DONE/"):
                    response = self.process_done(msg)

                elif msg.startswith("WINNERS/"):
                    response = self.process_winners(msg)

                else:
                    response, _dni, _num, ok = self.process_bet(msg)
                    logging.info(
                        f"action: apuesta_recibida | result: {'success' if ok else 'fail'} | cantidad: 1"
                    )

                response_bytes = response.encode("utf-8")
                client_sock.sendall(struct.pack(">I", len(response_bytes)) + response_bytes)
        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            try:
                client_sock.close()
            except Exception:
                pass

    def process_done(self, message: str) -> str:
        parts = message.split("/")
        if len(parts) != 2:
            return "RESPONSE/FAIL/DONE"

        agency_id = parts[1]

        # 1) Actualizar estado de DONE bajo lock
        should_draw = False
        with self._draw_cv:
            self._done_agencies.add(agency_id)

            # Si ya se sorteó, nada más que hacer
            if self._draw_done:
                return "RESPONSE/SUCCESS/DONE"

            # Si falta config, no podemos decidir “cuándo” sortear
            if self._required_agencies <= 0:
                return "RESPONSE/SUCCESS/DONE"

            # Si ya hay sorteo en curso, no disparar otro
            if self._drawing_in_progress:
                return "RESPONSE/SUCCESS/DONE"

            # Condición de disparo (solo un hilo la toma)
            if len(self._done_agencies) >= self._required_agencies:
                self._drawing_in_progress = True
                should_draw = True

        # 2) Computar ganadores FUERA del lock (evita bloquear WINNERS y otros DONE)
        if should_draw:
            winners = {}
            for bet in load_bets():
                if has_won(bet):
                    key = str(bet.agency)
                    winners.setdefault(key, []).append(str(bet.document))

            # 3) Publicar resultado del sorteo de forma atómica y despertar a los que esperen
            with self._draw_cv:
                # doble chequeo (por seguridad ante estados raros)
                if not self._draw_done:
                    self._winners_by_agency = winners
                    self._draw_done = True
                    logging.info("action: sorteo | result: success")

                self._drawing_in_progress = False
                self._draw_cv.notify_all()

        return "RESPONSE/SUCCESS/DONE"

    def process_winners(self, message: str) -> str:
        parts = message.split("/")
        if len(parts) != 2:
            return "RESPONSE/FAIL/WINNERS"

        agency_id = parts[1]

        # Si el sorteo está en curso, esperar a que termine (sin busy-wait)
        with self._draw_cv:
            while self._drawing_in_progress:
                self._draw_cv.wait()

            if not self._draw_done:
                return "RESPONSE/NOT_READY/WINNERS"

            dnis = self._winners_by_agency.get(agency_id, [])

        resp = f"RESPONSE/SUCCESS/WINNERS/{len(dnis)}"
        for dni in dnis:
            resp += f"/{dni}"
        return resp

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