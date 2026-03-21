package common

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"github.com/op/go-logging"
)

var protocolLog = logging.MustGetLogger("log")

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
	msgsize := uint16(msglen)
	sizeBufffer := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeBufffer, msgsize)

	if err := c.sendAll(sizeBufffer); err != nil {
		log.Errorf(
			"action: send_message | result: fail | error: %v")
		return err
	}

	if err := c.sendAll([]byte(message)); err != nil {
		log.Errorf(
			"action: send_message | result: fail | error: %v")
		return err
	}

	return nil

}

func (c *ClientProtocolMessage) createBetMessage(betInfo BetInfo) string {
	return fmt.Sprintf("BET/%s/%s/%s/%s/%s/%s\n", betInfo.Agency, betInfo.Name, betInfo.Surname, betInfo.DNI, betInfo.Birthday, betInfo.number)
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

	reader := bufio.NewReader(c.conn)
	var message []byte

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}

		if b == '\n' {
			return strings.TrimSpace(string(message)), nil
		}

		message = append(message, b)
	}
}

func (c *ClientProtocolMessage) reponseFromServer(response string) (*ServerResponse, error) {
	parts := strings.Split(response, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("Invalid server response format")
	}
	if parts[0] != "RESPONSE" {
		return nil, fmt.Errorf("Invalid server response type: %s", parts[0])
	}
	return &ServerResponse{
		Type:    parts[0],
		Status:  parts[1],
		Message: parts[2],
	}, nil
}
