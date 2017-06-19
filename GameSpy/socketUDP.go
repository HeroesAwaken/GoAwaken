package GameSpy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	log "github.com/ReviveNetwork/GoRevive/Log"
)

// Socket is a basic event-based TCP-Server
type SocketUDP struct {
	Clients   []*Client
	name      string
	port      string
	listen    *net.UDPConn
	eventChan chan SocketUDPEvent
	fesl      bool
}

type SocketUDPEvent struct {
	Name string
	Addr *net.UDPAddr
	Data interface{}
}

// New starts to listen on a new Socket
func (socket *SocketUDP) New(name string, port string, fesl bool) (chan SocketUDPEvent, error) {
	var err error

	socket.name = name
	socket.port = port
	socket.eventChan = make(chan SocketUDPEvent, 1000)
	socket.fesl = fesl

	// Listen for incoming connections.
	ServerAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:"+socket.port)
	if err != nil {
		log.Errorf("%s: Listening on 0.0.0.0:%s threw an error.\n%v", socket.name, socket.port, err)
		return nil, err
	}

	socket.listen, err = net.ListenUDP("udp", ServerAddr)
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
func (socket *SocketUDP) Close() {
	// Fire closing event
	log.Noteln(socket.name + " closing. Port " + socket.port)
	socket.eventChan <- SocketUDPEvent{
		Name: "close",
		Addr: nil,
		Data: nil,
	}

	// Close socket
	socket.listen.Close()
}

func (socket *SocketUDP) readFESL(data []byte, addr *net.UDPAddr) {
	outCommand := new(CommandFESL)

	p := bytes.NewBuffer(data)
	var payloadId uint32
	var payloadLen uint32

	payloadType := string(data[:4])
	p.Next(4)

	binary.Read(p, binary.BigEndian, &payloadId)
	binary.Read(p, binary.BigEndian, &payloadLen)

	payloadRaw := data[12:]
	payload := ProcessFESL(string(payloadRaw))

	outCommand.Query = payloadType
	outCommand.PayloadID = payloadId
	outCommand.Message = payload

	socket.eventChan <- SocketUDPEvent{
		Name: "command." + payloadType,
		Addr: addr,
		Data: outCommand,
	}
	socket.eventChan <- SocketUDPEvent{
		Name: "command",
		Addr: addr,
		Data: outCommand,
	}

}

func (socket *SocketUDP) processCommand(command string, addr *net.UDPAddr) {
	gsPacket, err := ProcessCommand(command)
	if err != nil {
		log.Errorf("%s: Error processing command %s.\n%v", socket.name, command, err)
		socket.eventChan <- SocketUDPEvent{
			Name: "error",
			Addr: addr,
			Data: err,
		}
		return
	}

	socket.eventChan <- SocketUDPEvent{
		Name: "command." + gsPacket.Query,
		Addr: addr,
		Data: gsPacket,
	}
	socket.eventChan <- SocketUDPEvent{
		Name: "command",
		Addr: addr,
		Data: gsPacket,
	}
}

func (socket *SocketUDP) run() {
	buf := make([]byte, 4096)

	for {
		n, addr, err := socket.listen.ReadFromUDP(buf)
		if err != nil {
			log.Errorf("%s: Error reading from UDP.%v", socket.name, err)
			socket.eventChan <- SocketUDPEvent{
				Name: "error",
				Addr: addr,
				Data: err,
			}
			continue
		}

		if socket.fesl {
			socket.readFESL(buf[:n], addr)
			continue
		}

		message := strings.TrimSpace(string(socket.XOr(buf[0:n])))

		log.Debugln("Got UDP message:", message)

		socket.eventChan <- SocketUDPEvent{
			Name: "data",
			Addr: addr,
			Data: message,
		}

		socket.processCommand(message, addr)
	}
}

func (socket *SocketUDP) WriteFESL(msgType string, msg map[string]string, msgType2 uint32, addr *net.UDPAddr) error {
	var lena int32
	var buf bytes.Buffer

	payloadEncoded := SerializeFESL(msg)
	baselen := len(payloadEncoded)
	lena = int32(baselen + 12)

	buf.Write([]byte(msgType))

	err := binary.Write(&buf, binary.BigEndian, &msgType2)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}

	err = binary.Write(&buf, binary.BigEndian, &lena)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}

	buf.Write([]byte(payloadEncoded))

	log.Debugln("Write message:", msg, msgType, msgType2)

	n, err := socket.listen.WriteToUDP(buf.Bytes(), addr)
	if err != nil {
		fmt.Println("Writing failed:", n, err)
	}
	return nil
}

func (socket *SocketUDP) Write(message string, addr *net.UDPAddr) {
	log.Debugln("Sending message:", message)
	xOrMessage := socket.XOr([]byte(message))

	_, err := socket.listen.WriteToUDP(xOrMessage, addr)
	if err != nil {
		log.Errorf("%s: Error writing to UDP. Message:%s Client:%v %v", socket.name, message, addr, err)
		socket.eventChan <- SocketUDPEvent{
			Name: "error",
			Addr: addr,
			Data: err,
		}
	}
}

// XOr applies the gamespy XOr
func (socket *SocketUDP) XOr(a []byte) []byte {
	b := []byte("gamespy")
	var res []byte

	var k = 0
	var i = 0
	for i < len(a) {
		if k > (len(b) - 1) {
			k = 0
		}
		res = append(res, a[i]^b[k])
		k++
		i++
	}

	return res
}
