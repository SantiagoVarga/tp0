import logging

class Protocol:
    
    def __init__(self,socket):
        self.socket = socket

    def receive_message(self) -> str:
        msgSize = int.from_bytes(self.socket.recv(2), byteorder='big')

        message = b''
        while len(message) < msgSize:
            chunk = self.socket.recv(msgSize - len(message))
            if not chunk:
                raise ConnectionError("Connection closed by the server while receiving message.")
            message += chunk
        return message.decode('utf-8').strip()

    def send_message(self, message: str):
        data = message.encode('utf-8')
        total_sent = 0
        while total_sent < len(data):
            sent = self.socket.send(data[total_sent:])
            if sent == 0:
                raise ConnectionError("Connection closed by the server while sending message.")
            total_sent += sent

    def close(self):
        try:
            self.socket.close()
        except Exception as e:
            logging.error("action: close_socket | result: fail | error: {e}")