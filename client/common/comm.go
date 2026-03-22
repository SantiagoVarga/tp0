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
	return fmt.Sprintf("BET/%s/%s/%s/%s/%s/%s", betInfo.Agency, betInfo.Name, betInfo.Surname, betInfo.DNI, betInfo.Birthday, betInfo.Number)
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
	// Espera un string tipo RESPONSE/SUCCESS/Apuesta almacenada o RESPONSE/FAIL/Error al almacenar apuesta
	parts := strings.SplitN(response, "/", 3)
	if len(parts) != 3 {
		return &ServerResponse{
			Type:    "RESPONSE",
			Status:  "FAIL",
			Message: "Formato de respuesta inválido",
		}, nil
	}
	status := "FAIL"
	if parts[1] == "SUCCESS" {
		status = "SUCCESS"
	}
	return &ServerResponse{
		Type:    parts[0],
		Status:  status,
		Message: parts[2],
	}, nil
}

// SerializeBatch serializa un batch de apuestas con prefijo BET/BATCH
func (c *ClientProtocolMessage) SerializeBatch(bets []BetInfo) string {
	message := fmt.Sprintf("BET/BATCH/%d", len(bets))
	for _, bet := range bets {
		message += fmt.Sprintf("/%s/%s/%s/%s/%s/%s", bet.Agency, bet.Name, bet.Surname, bet.DNI, bet.Birthday, bet.Number)
	}
	return message
}

// SendBatch envía un batch de apuestas al servidor
func (c *ClientProtocolMessage) SendBatch(bets []BetInfo) (string, error) {
	message := c.SerializeBatch(bets)
	err := c.sendAll([]byte(message))
	if err != nil {
		return "", err
	}

	response, err := c.receiveMessage()
	if err != nil {
		return "", err
	}

	return response, nil
}
