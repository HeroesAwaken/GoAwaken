package GameSpy

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strings"

	log "github.com/ReviveNetwork/GoRevive/Log"
)

type Client struct {
	name       string
	conn       *net.Conn
	recvBuffer []byte
	eventChan  chan ClientEvent
	IsActive   bool
	reader     *bufio.Reader
	State      interface{}
}

// ClientEvent is the generic struct for events
// by this Client
type ClientEvent struct {
	Name string
	Data interface{}
}

// New creates a new Client and starts up the handling of the connection
func (client *Client) New(name string, conn *net.Conn) (chan ClientEvent, error) {
	client.name = name
	client.conn = conn
	client.eventChan = make(chan ClientEvent, 1000)
	client.reader = bufio.NewReader(*client.conn)

	go client.handleRequest()

	return client.eventChan, nil
}

func (client *Client) Write(command string) error {
	if !client.IsActive {
		log.Notef("%s: Trying to write to inactive client.\n%v", client.name, command)
		return errors.New("client is not active. Can't send message")
	}

	(*client.conn).Write([]byte(command))
	return nil
}

// WriteError Handy for informing the user they're a piece of shit.
func (client *Client) WriteError(code string, message string) error {
	err := client.Write("\\error\\\\err\\" + code + "\\fatal\\\\errmsg\\" + message + "\\id\\1\\final\\")
	return err
}

func (client *Client) processCommand(command string) {
	gsPacket, err := ProcessCommand(command)
	if err != nil {
		log.Errorf("%s: Error processing command %s.\n%v", client.name, command, err)
		client.eventChan <- ClientEvent{
			Name: "error",
			Data: err,
		}
		return
	}

	client.eventChan <- ClientEvent{
		Name: "command." + gsPacket.Query,
		Data: gsPacket,
	}
	client.eventChan <- ClientEvent{
		Name: "command",
		Data: gsPacket,
	}
}

func (client *Client) handleRequest() {
	var err error
	client.IsActive = true

	for client.IsActive {
		// Make a buffer to hold incoming data.
		for {
			client.recvBuffer, err = client.reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					log.Errorf("%s: Reading from client threw an error.\n%v", client.name, err)
					client.eventChan <- ClientEvent{
						Name: "error",
						Data: err,
					}
				} else {
					// If we receive an EndOfFile, close this function/goroutine
					log.Notef("%s: Client closing connection.", client.name)
					client.eventChan <- ClientEvent{
						Name: "close",
						Data: client,
					}
					client.IsActive = false
					return
				}
				break
			}

			// Send a response back to person contacting us.
			(*client.conn).Write([]byte("Message received.\n"))

			message := strings.TrimSpace(string(client.recvBuffer))

			client.eventChan <- ClientEvent{
				Name: "data",
				Data: message,
			}

			if strings.Index(message, "\\final\\") == -1 {
				continue
			}

			for _, command := range strings.Split(message, "\\final\\") {
				if len(command) == 0 {
					break
				}

				client.processCommand(command)
			}
		}
	}

}
