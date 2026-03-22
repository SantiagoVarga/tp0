package common

import (
	"net"
	"strings"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config      ClientConfig
	conn        net.Conn
	Bets        []BetInfo
	protocol    ClientProtocolMessage
	BatchConfig BatchConfig
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	batchIndex := 0

	for batchIndex < len(c.Bets) {
		end := batchIndex + c.BatchConfig.MaxAmount
		if end > len(c.Bets) {
			end = len(c.Bets)
		}

		batch := c.Bets[batchIndex:end]
		batchSize := end - batchIndex

		response, err := c.protocol.SendBatch(batch)
		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | cantidad: %d | error: %v", batchSize, err)
			batchIndex = end
			continue
		}

		if c.handleServerResponse(response, batchSize) {
			log.Infof("action: apuesta_enviada | result: success | cantidad: %d", batchSize)
		} else {
			log.Errorf("action: apuesta_enviada | result: fail | cantidad: %d", batchSize)
		}

		batchIndex = end
		time.Sleep(c.config.LoopPeriod)
	}
}

// CloseResources Graceful shutdown for client
func (c *Client) CloseResources() {
	log.Infof("Cerrando conexión del cliente...")
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			log.Errorf("Error al cerrar conexión: %v", err)
		} else {
			log.Infof("Conexión cerrada correctamente.")
		}
		c.conn = nil
	}
}

func (c *Client) handleServerResponse(response string, batchSize int) bool {
	parts := strings.SplitN(response, "/", 3)
	if len(parts) < 3 {
		return false
	}

	prefix := parts[0]
	status := parts[1]

	if prefix == "RESPONSE" && status == "SUCCESS" {
		return true
	}

	return false
}
