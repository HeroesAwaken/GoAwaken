package GameSpy

import (
	"crypto/tls"
	"errors"
	"net"
	"strings"

	log "github.com/ReviveNetwork/GoRevive/Log"
)

// Socket is a basic event-based TCP-Server
type SocketTLS struct {
	ClientsTLS []*ClientTLS
	name       string
	port       string
	listen     net.Listener
	eventChan  chan SocketEvent
}

type EventNewClientTLS struct {
	Client *ClientTLS
}

type EventClientTLSClose struct {
	Client *ClientTLS
}
type EventClientTLSError struct {
	Client *ClientTLS
	Error  error
}
type EventClientTLSCommand struct {
	Client  *ClientTLS
	Command *Command
}
type EventClientTLSData struct {
	Client *ClientTLS
	Data   string
}

// New starts to listen on a new Socket
func (socket SocketTLS) New(name string, port string, tlsCert string, tlsKey string) (chan SocketEvent, error) {
	var err error

	socket.name = name
	socket.port = port
	socket.eventChan = make(chan SocketEvent, 1000)

	// Listen for incoming connections.
	cer, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates:       []tls.Certificate{cer},
		ClientAuth:         tls.NoClientCert,
		MinVersion:         tls.VersionSSL30,
		InsecureSkipVerify: true,
		//MaxVersion:   tls.VersionSSL30,
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_RC4_128_SHA,
			0x014,
		},
	}
	socket.listen, err = tls.Listen("tcp", "0.0.0.0:"+socket.port, config)

	if err != nil {
		log.Errorf("%s: Listening on 0.0.0.0:%s threw an error.\n%v", socket.name, socket.port, err)
		return nil, err
	}
	log.Noteln(socket.name + ": Listening on 0.0.0.0:" + socket.port)

	// Accept new connections in a new Goroutine("thread")
	go socket.run()

	return socket.eventChan, nil
}

// Close fires a close-event and closes the socket
func (socket SocketTLS) Close() {
	// Fire closing event
	log.Noteln(socket.name + " closing. Port " + socket.port)
	socket.eventChan <- SocketEvent{
		Name: "close",
		Data: nil,
	}

	// Close socket
	socket.listen.Close()
}

func (socket SocketTLS) run() {
	for {
		// Listen for an incoming connection.
		conn, err := socket.listen.Accept()
		if err != nil {
			log.Errorf("%s: A new client connecting threw an error.\n%v", socket.name, err)
			socket.eventChan <- SocketEvent{
				Name: "error",
				Data: EventError{
					Error: err,
				},
			}
			continue
		}

		tlscon, ok := conn.(*tls.Conn)
		if !ok {
			log.Errorf("%s: A new client connecting threw an error.\n%v", socket.name, err)
			socket.eventChan <- SocketEvent{
				Name: "error",
				Data: EventError{
					Error: err,
				},
			}
			continue
		}

		/*state := tlscon.ConnectionState()
		log.Debugf("Connection handshake complete %v, %v", state.HandshakeComplete, state)

		err = tlscon.Handshake()
		if err != nil {
			log.Errorf("%s: A new client connecting threw an error.\n%v", socket.name, err)
			socket.eventChan <- SocketEvent{
				Name: "error",
				Data: EventError{
					Error: err,
				},
			}
			continue
		}*/

		// Create a new Client and add it to our slice
		newClient := new(ClientTLS)
		newClient.FESL = true
		clientEventSocket, err := newClient.New(socket.name, tlscon)
		if err != nil {
			log.Errorf("%s: Creating the new client threw an error.\n%v", socket.name, err)
			socket.eventChan <- SocketEvent{
				Name: "error",
				Data: EventError{
					Error: err,
				},
			}
		}
		go socket.handleClientEvents(newClient, clientEventSocket)

		log.Noteln(socket.name + ": A new client connected")
		socket.ClientsTLS = append(socket.ClientsTLS, newClient)

		// Fire newClient event
		socket.eventChan <- SocketEvent{
			Name: "newClient",
			Data: EventNewClientTLS{
				Client: newClient,
			},
		}
	}
}

func (socket SocketTLS) removeClient(client *ClientTLS) error {
	var indexToRemove = 0
	var foundClient = false

	log.Debugln("Removing client ", client)

	client.IsActive = false
	(*client.conn).Close()

	for i := range socket.ClientsTLS {
		if socket.ClientsTLS[i] == client {
			indexToRemove = i
			foundClient = true
			break
		}
	}

	if !foundClient {
		return errors.New("could not find client to remove")
	}

	log.Debugln("Found client as ", indexToRemove)

	if len(socket.ClientsTLS) == 1 {
		// We have only one element, so create a new one
		socket.ClientsTLS = []*ClientTLS{}
		return nil
	}

	// Replace our client set to remove with the last client in the array
	// and then cut the last element of the array
	socket.ClientsTLS[indexToRemove] = socket.ClientsTLS[len(socket.ClientsTLS)-1]
	socket.ClientsTLS = socket.ClientsTLS[:len(socket.ClientsTLS)-1]

	log.Debugln("Client removed")
	return nil
}

func (socket SocketTLS) handleClientEvents(client *ClientTLS, eventsChannel chan ClientTLSEvent) {
	for client.IsActive {
		select {
		case event := <-eventsChannel:
			switch {
			case event.Name == "close":
				socket.eventChan <- SocketEvent{
					Name: "client." + event.Name,
					Data: EventClientTLSClose{
						Client: client,
					},
				}
				err := socket.removeClient(client)
				if err != nil {
					log.Errorln("Could not remove client", err)
				}
			case strings.Index(event.Name, "command") != -1:
				socket.eventChan <- SocketEvent{
					Name: "client." + event.Name,
					Data: EventClientTLSCommand{
						Client:  client,
						Command: event.Data.(*Command),
					},
				}
			case event.Name == "data":
				socket.eventChan <- SocketEvent{
					Name: "client." + event.Name,
					Data: EventClientTLSData{
						Client: client,
						Data:   event.Data.(string),
					},
				}

			default:
				var interfaceSlice = make([]interface{}, 2)
				interfaceSlice[0] = client
				interfaceSlice[1] = event.Data

				// Send the event down the chain
				socket.eventChan <- SocketEvent{
					Name: "client." + event.Name,
					Data: interfaceSlice,
				}
			}
			/*default:
			if !client.IsActive {
				break
			}
			runtime.Gosched()*/
		}
	}

	socket.removeClient(client)
}
