package common

import (
	"net"
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
	config   ClientConfig
	conn     net.Conn
	BetInfo  BetInfo
	protocol ClientProtocolMessage
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
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		c.protocol = NewClientProtocolMessage(c.conn, c.config.ID)

		response, err := c.protocol.sendBet(c.BetInfo)

		if err != nil {
			log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		c.handleServerResponse(response)

		// TODO: Modify the send to avoid short-write
		/*fmt.Fprintf(
			c.conn,
			"[CLIENT %v] Message N°%v\n",
			c.config.ID,
			msgID,
		)
		msg, err := bufio.NewReader(c.conn).ReadString('\n')*/
		c.conn.Close()

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		/*log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)*/

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
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

func (c *Client) handleServerResponse(response *ServerResponse) {
	if response.Status == "SUCCESS" {
		log.Infof("action: apuesta_enviada | result: success | dni: %s | numero: %s", c.BetInfo.DNI, c.BetInfo.number)
	} else {
		log.Errorf("action: apuesta_enviada | result: fail | dni: %s | numero: %s | error: %s", c.BetInfo.DNI, c.BetInfo.number, response.Message)
	}
}
