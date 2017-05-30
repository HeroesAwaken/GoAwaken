package GameSpy

import (
	"net"
)

type Client struct {
	name       string
	conn       *net.Conn
	recvBuffer []byte
	eventChan  chan ClientEvent
}

// ClientEvent is the generic struct for events
// by this Client
type ClientEvent struct {
	Name string
	Data interface{}
}

func (client *Client) New(name string, conn *net.Conn) (chan ClientEvent, error) {
	client.name = name
	client.conn = conn
	client.eventChan = make(chan ClientEvent)

	return client.eventChan, nil
}
