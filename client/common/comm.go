package common

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
)

type ClientProtocolMessage struct {
	conn net.Conn
	id   string
}

func NewClientProtocolMessage(conn net.Conn, id string) ClientProtocolMessage {
	return ClientProtocolMessage{conn: conn, id: id}
}

func (c *ClientProtocolMessage) sendAll(message []byte) error {
	totalSent := 0
	for totalSent < len(message) {
		n, err := c.conn.Write(message[totalSent:])
		if err != nil {
			return err
		}
		totalSent += n
	}
	return nil
}

func (c *ClientProtocolMessage) sendMessage(message string) error {
	if c.conn == nil {
		return fmt.Errorf("Connection failed")
	}
	msglen := len(message)
	sizeBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuffer, uint32(msglen))

	if err := c.sendAll(sizeBuffer); err != nil {
		log.Errorf("action: send_message | result: fail | error: %v", err)
		return err
	}

	if err := c.sendAll([]byte(message)); err != nil {
		log.Errorf("action: send_message | result: fail | error: %v", err)
		return err
	}

	return nil
}

func (c *ClientProtocolMessage) createBetMessage(betInfo BetInfo) string {
	// Prefijo BET y campos separados por /
	return fmt.Sprintf("BET/%s/%s/%s/%s/%s/%s", betInfo.Agency, betInfo.Name, betInfo.Surname, betInfo.DNI, betInfo.Birthday, betInfo.number)
}

func (c *ClientProtocolMessage) sendBet(betInfo BetInfo) (*ServerResponse, error) {
	message := c.createBetMessage(betInfo)

	if err := c.sendMessage(message); err != nil {
		return nil, fmt.Errorf("Failed to send bet message: %v", err)
	}

	responseStr, err := c.receiveMessage()
	if err != nil {
		return nil, fmt.Errorf("Failed to receive response: %v", err)
	}

	response, err := c.reponseFromServer(responseStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse server response: %v", err)
	}

	return response, nil

}

func (c *ClientProtocolMessage) receiveMessage() (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("Connection failed")
	}
	// Leer 4 bytes de longitud
	sizeBuffer := make([]byte, 4)
	if _, err := io.ReadFull(c.conn, sizeBuffer); err != nil {
		return "", fmt.Errorf("Failed to read response size: %v", err)
	}
	msglen := binary.BigEndian.Uint32(sizeBuffer)
	// Leer el mensaje completo
	msgBuffer := make([]byte, msglen)
	if _, err := io.ReadFull(c.conn, msgBuffer); err != nil {
		return "", fmt.Errorf("Failed to read response body: %v", err)
	}
	return string(msgBuffer), nil
}

func (c *ClientProtocolMessage) reponseFromServer(response string) (*ServerResponse, error) {
	// Espera un JSON tipo {"status": "ok"} o {"status": "fail"}
	status := ""
	message := ""
	if strings.Contains(response, "ok") {
		status = "SUCCESS"
		message = "Apuesta almacenada"
	} else {
		status = "FAIL"
		message = "Error al almacenar apuesta"
	}
	return &ServerResponse{
		Type:    "RESPONSE",
		Status:  status,
		Message: message,
	}, nil
}
