package nanomsg

import (
	"bytes"
	"fmt"
	"time"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pull"
	"github.com/go-mangos/mangos/transport/tcp"
	"github.com/joernweissenborn/eventual2go"
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
	socket         mangos.Socket
	stop           *eventual2go.Completer
}

// Init initializes a providers incoming channel.
func (p *Provider) Init(cfg *config.Config) (err error) {

	if p.socket, err = pull.NewSocket(); err != nil {
		return
	}

	p.socket.AddTransport(tcp.NewTransport())

	if err = p.socket.SetOption("RECV-DEADLINE", 100*time.Millisecond); err != nil {
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

	if err = p.socket.Listen(fmt.Sprintf("tcp://%s:%d", iface, port)); err != nil {
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

	for !stop.Completed() {
		msg, err := p.socket.RecvMsg()
		if err != nil && err != mangos.ErrRecvTimeout {
			panic(err)
		} else if err == nil {
			p.messages.Add(msg)
		}
	}
}

func (p *Provider) Shutdown(eventual2go.Data) (err error) {
	err = p.socket.Close()
	return
}

func decode(d eventual2go.Data) eventual2go.Data {
	m := d.(*mangos.Message)
	dec := codec.NewDecoder(bytes.NewBuffer(m.Body), &mh)

	var msg network.Message
	dec.Decode(&msg)

	return msg
}
