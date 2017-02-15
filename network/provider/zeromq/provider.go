package zeromq

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/joernweissenborn/eventual2go"
	"github.com/pebbe/zmq4"
	"github.com/ugorji/go/codec"
)

var (
	mh codec.MsgpackHandle
)

type Config struct {
	Address string
	Port    int
}

type Provider struct {
	config         *config.Config
	providerConfig Config
	messages       *eventual2go.StreamController
	socket         *zmq4.Socket
	stop           *eventual2go.Completer
}

// Init initializes a providers incoming channel.
func (p *Provider) Init(cfg *config.Config) (err error) {

	if p.socket, err = zmq4.NewSocket(zmq4.ROUTER); err != nil {
		return
	}

	iface := cfg.User.Interfaces[0]

	port, err := network.GetFreePortOnInterface(iface)
	if err != nil {
		return
	}

	p.providerConfig = Config{
		Address: iface,
		Port:    port,
	}

	if err = p.socket.Bind(fmt.Sprintf("tcp://%s:%d", iface, port)); err != nil {
		return
	}

	p.config = cfg

	p.messages = eventual2go.NewStreamController()

	p.stop = eventual2go.NewCompleter()

	go p.receive()

	return
}

// Connect connectes to peer using the given details.
func (p *Provider) Connect(details network.Details, uuid uuid.UUID) (conn network.Connection, err error) {
	var cfg Config
	dec := codec.NewDecoder(bytes.NewBuffer(details.Config), &mh)
	dec.Decode(&cfg)

	ams, err := eventual2go.SpawnActor(&Connection{
		address: cfg.Address,
		port:    cfg.Port,
		cfg:     p.config,
	})
	conn = network.Connection{
		ActorMessageStream: ams,
		UUID:               uuid,
	}

	return
}

// Details returns the details of the incoming connection. This will be advertised to other peers.
func (p *Provider) Details() (details network.Details) {
	var cfg bytes.Buffer
	enc := codec.NewEncoder(&cfg, &mh)
	enc.Encode(p.providerConfig)
	details = network.Details{
		Provider: network.NANOMSG,
		Config:   cfg.Bytes(),
	}
	return
}

// Messages returns a stream of incoming messages.
func (p *Provider) Messages() *network.MessageStream {
	return &network.MessageStream{Stream: p.messages.Stream().Transform(decode)}
}

func (p *Provider) receive() {

	stop := p.stop.Future()
	poller := zmq4.NewPoller()
	poller.Add(p.socket, zmq4.POLLIN)

	for !stop.Completed() {
		sockets, err := poller.Poll(100 * time.Millisecond)
		if err != nil {
			continue
		}
		for range sockets {
			msg, err := p.socket.RecvMessageBytes(0)
			if err == nil {
				p.messages.Add(msg)
			}
		}
	}
}

func (p *Provider) Shutdown(eventual2go.Data) (err error) {
	p.stop.Complete(nil)
	err = p.socket.Close()
	return
}

func decode(d eventual2go.Data) eventual2go.Data {
	m := d.([][]byte)

	msgType, _ := binary.Varint(m[1])

	msg := network.Message{
		Sender:  uuid.UUID(m[0]),
		Type:    message.Type(msgType),
		Payload: m[2:],
	}
	return msg
}
