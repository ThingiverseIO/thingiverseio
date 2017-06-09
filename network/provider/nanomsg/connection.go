package nanomsg

import (
	"bytes"
	"fmt"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/push"
	"github.com/go-mangos/mangos/transport/tcp"
	"github.com/joernweissenborn/eventual2go"
	"github.com/ugorji/go/codec"
)

type Connection struct {
	address string
	port    int
	cfg     *config.Config
	socket  mangos.Socket
}

func (c *Connection) Init() (err error) {

	if c.socket, err = push.NewSocket(); err != nil {
		return
	}

	c.socket.AddTransport(tcp.NewTransport())

	err = c.socket.Dial(fmt.Sprintf("tcp://%s:%d", c.address, c.port))

	return
}

func (c *Connection) OnMessage(d eventual2go.Data) {
	msg := d.(network.Message)
	msg.Sender = c.cfg.Internal.UUID
	var buf bytes.Buffer
	enc := codec.NewEncoder(&buf, &mh)
	enc.Encode(msg)
	body := buf.Bytes()

	m := mangos.NewMessage(len(body))
	m.Body = body
	c.socket.SendMsg(m)
}

func (c *Connection) Shutdown(d eventual2go.Data) error {
	return c.socket.Close()
}
