package GameSpy

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strings"
	"time"

	log "github.com/ReviveNetwork/GoRevive/Log"
)

type Client struct {
	name       string
	conn       *net.Conn
	recvBuffer []byte
	eventChan  chan ClientEvent
	IsActive   bool
	reader     *bufio.Reader
	IpAddr     net.Addr
	State      ClientState
}

type ClientState struct {
	ServerChallenge string
	ClientChallenge string
	ClientResponse  string
	BattlelogID     int
	Username        string
	PlyName         string
	PlyEmail        string
	PlyCountry      string
	PlyPid          int
	Confirmed       bool
	Banned          bool
	IpAddress       net.Addr
	HasLogin        bool
	ProfileSent     bool
	LoggedOut       bool
	HeartTicker     *time.Ticker
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
	client.IpAddr = (*client.conn).RemoteAddr()
	client.eventChan = make(chan ClientEvent, 1000)
	client.reader = bufio.NewReader(*client.conn)
	client.IsActive = true

	go client.handleRequest()

	return client.eventChan, nil
}

func (client *Client) Write(command string) error {
	if !client.IsActive {
		log.Notef("%s: Trying to write to inactive client.\n%v", client.name, command)
		return errors.New("client is not active. Can't send message")
	}

	log.Debugln("Write message:", command)

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

func (client *Client) Close() {
	log.Notef("%s: Client closing connection.", client.name)
	client.eventChan <- ClientEvent{
		Name: "close",
		Data: client,
	}
	client.IsActive = false
}

func (client *Client) handleRequest() {
	client.IsActive = true

	for client.IsActive {
		// Make a buffer to hold incoming data.
		buf := make([]byte, 4096) // buffer
		for {
			n, err := (*client.conn).Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Debugf("%s: Reading from client threw an error. %v", client.name, err)
					client.eventChan <- ClientEvent{
						Name: "error",
						Data: err,
					}
					client.eventChan <- ClientEvent{
						Name: "close",
						Data: client,
					}
					client.IsActive = false
					return
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

			client.recvBuffer = append(client.recvBuffer, buf[:n]...)

			message := strings.TrimSpace(string(client.recvBuffer))

			if strings.Index(message, "\\final\\") == -1 {
				continue
			}

			log.Debugln("Got message:", message)

			client.eventChan <- ClientEvent{
				Name: "data",
				Data: message,
			}

			commands := strings.Split(message, "\\final\\")
			for _, command := range commands {
				if len(command) == 0 {
					break
				}

				client.processCommand(command)
			}

			// Add unprocessed commands back into recvBuffer
			client.recvBuffer = []byte(commands[(len(commands) - 1)])
		}
	}

}
