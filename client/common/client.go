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
		return err
	}
	c.conn = conn
	c.protocol = NewClientProtocolMessage(c.conn, c.config.ID)
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	if err := c.createClientSocket(); err != nil {
		return
	}
	defer c.CloseResources()

	maxBetsPerBatch := c.BatchConfig.MaxAmount
	if maxBetsPerBatch <= 0 {
		maxBetsPerBatch = 99
	}

	for batchStart := 0; batchStart < len(c.Bets); batchStart += maxBetsPerBatch {
		batchEnd := batchStart + maxBetsPerBatch
		if batchEnd > len(c.Bets) {
			batchEnd = len(c.Bets)
		}

		batch := c.Bets[batchStart:batchEnd]
		batchSize := len(batch)
		response, err := c.protocol.SendBatch(batch)
		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | cantidad: %d | error: %v", batchSize, err)
			continue
		}

		if c.handleServerResponse(response) {
			log.Infof("action: apuesta_enviada | result: success | cantidad: %d", batchSize)
		} else {
			log.Errorf("action: apuesta_enviada | result: fail | cantidad: %d", batchSize)
		}

		time.Sleep(c.config.LoopPeriod)
	}

	_, _ = c.protocol.NotifyDone(c.config.ID)

	c.CloseResources()

	for {
		for {
			if err := c.createClientSocket(); err == nil {
				break
			}
			time.Sleep(150 * time.Millisecond)
		}

		wr, err := c.protocol.RequestWinners(c.config.ID)
		c.CloseResources()

		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if !wr.Ready {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", wr.Count)
		return
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

func (c *Client) handleServerResponse(response string) bool {
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
