package zeromq

import (
	"encoding/binary"
	"fmt"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/joernweissenborn/eventual2go"
	"github.com/pebbe/zmq4"
)

type Connection struct {
	address string
	port    int
	cfg     *config.Config
	socket  *zmq4.Socket
}

func (c *Connection) Init() (err error) {

	if c.socket, err = zmq4.NewSocket(zmq4.DEALER); err != nil {
		return
	}

	if err = c.socket.SetIdentity(string(c.cfg.Internal.UUID)); err != nil {
		return
	}

	err = c.socket.Connect(fmt.Sprintf("tcp://%s:%d", c.address, c.port))

	return
}

func (c *Connection) OnMessage(d eventual2go.Data) {
	msg := d.(network.Message)

	size := binary.Size(int64(msg.Type))
	msgType := make([]byte, size)
	binary.PutVarint(msgType, int64(msg.Type))

	c.socket.SendMessage(append([][]byte{msgType}, msg.Payload...))
}

func (c *Connection) Shutdown(d eventual2go.Data) {
	c.socket.Close()
}
