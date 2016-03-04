package peer

import (
	"fmt"
	"log"
	"time"

	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/thingiverseio/config"
	"github.com/joernweissenborn/thingiverseio/service/connection"
	"github.com/joernweissenborn/thingiverseio/service/messages"
)

//go:generate event_generator -t *Peer -n Peer

// Peer is node with an rpc connection.
type Peer struct {
	uuid        config.UUID
	cfg         *config.Config
	incoming    *connection.Incoming
	initialized *eventual2go.Completer
	connected   *PeerCompleter
	msgIn       *messages.MessageStream
	msgOut      *connection.Outgoing
	removed     *PeerCompleter
	logger      *log.Logger
}

// New creates a new Peer.
func New(uuid config.UUID, address string, port int, incoming *connection.Incoming, cfg *config.Config) (p *Peer, err error) {
	p = &Peer{
		uuid:      uuid,
		cfg:       cfg,
		incoming:  incoming,
		msgIn:     incoming.MessagesFromSender(uuid),
		connected: NewPeerCompleter(),
		removed:   NewPeerCompleter(),
		logger:    log.New(cfg.Logger(), fmt.Sprintf("%s PEER %s ", cfg.UUID(), uuid), 0),
	}

	p.msgOut, err = connection.NewOutgoing(cfg.UUID(), address, port)
	if err != nil {
		return
	}

	p.msgIn.Where(messages.Is(messages.HELLOOK)).Listen(p.onHelloOk)
	p.msgIn.Where(messages.Is(messages.END)).Listen(p.onEnd)

	if cfg.Exporting() {
		p.msgIn.Where(messages.Is(messages.DOHAVE)).Listen(p.onDoHave)
		p.msgIn.Where(messages.Is(messages.CONNECT)).Listen(p.onConnected)
	}

	p.msgIn.CloseOnFuture(p.removed.Future().Future)
	p.removed.Future().Then(p.closeOutgoing)

	p.initialized = eventual2go.NewTimeoutCompleter(5 * time.Second)
	p.initialized.Future().Err(p.onError)
	return
}

// NewFromHello creates a new Peer from a HELLO message.
func NewFromHello(m *messages.Hello, incoming *connection.Incoming, cfg *config.Config) (p *Peer, err error) {
	p, err = New(config.UUID(m.UUID), m.Address, m.Port, incoming, cfg)
	if err != nil {
		return
	}
	p.logger.Println("Received HELLO")

	p.Send(&messages.HelloOk{})

	return
}

// InitConnection initializes the connection.
func (p *Peer) InitConnection() {

	if p.initialized.Completed() {
		return
	}

	hello := &messages.Hello{string(p.cfg.UUID()), p.incoming.Addr(), p.incoming.Port()}

	p.Send(hello)

}

func (p *Peer) UUID() config.UUID {
	return p.uuid
}

// Messages returns the peers message stream.
func (p *Peer) Messages() *messages.MessageStream {
	return p.msgIn
}

// Connected returns a future which gets completed if the peer succesfully connects.
func (p *Peer) Connected() *PeerFuture {
	return p.connected.Future()
}

// Remove closes the peer.
func (p *Peer) Remove() {
	p.logger.Println("Removing")
	if !p.removed.Completed() {
		p.Send(&messages.End{})
		p.removed.Complete(p)
	}
}

func (p *Peer) onError(err error) (eventual2go.Data, error) {
	p.logger.Println("ERROR", err)
	p.Remove()
	return nil, nil
}

func (p *Peer) onEnd(messages.Message) {
	p.logger.Println("Received END")
	p.Remove()
}
func (p *Peer) onHelloOk(messages.Message) {
	p.logger.Println("Received HELLO_OK")
	if !p.initialized.Completed() {
		p.logger.Println("Initialized")
		p.initialized.Complete(nil)
		p.Send(&messages.HelloOk{})
	}
}

func (p *Peer) onDoHave(m messages.Message) {

	p.logger.Println("Received DOHAVE")
	if p.cfg.Exporting() {
		dohave := m.(*messages.DoHave)
		v, have := p.cfg.Tags()[dohave.TagKey]
		if have {
			have = v == dohave.TagValue
		}
		p.logger.Println("Checking tag", dohave, have)
		p.Send(&messages.Have{have, dohave.TagKey, dohave.TagValue})
		if !have {
			p.Remove()
		}
	}
}

func (p *Peer) onConnected(m messages.Message) {
	p.logger.Println("Received CONNECTED")
	p.connected.Complete(p)
}

// Removed returns a future which is completed when the connection gets closed.
func (p *Peer) Removed() *PeerFuture {
	return p.removed.Future()
}

// Send send a message to the peer.
func (p *Peer) Send(m messages.Message) {
	p.logger.Println("Sending Message ", m.GetType())
	p.msgOut.Send(messages.Flatten(m))
}

func (p *Peer) closeOutgoing(*Peer) *Peer {
	p.msgOut.Close()
	return nil
}

// Check sends all tags to the peer and gets an answer if the peer supports it.
func (p *Peer) Check() {
	p.initialized.Future().Then(p.check)
}

func (p *Peer) check(eventual2go.Data) eventual2go.Data {
	msgs := p.Messages().Where(messages.Is(messages.HAVE))
	c := msgs.AsChan()
	p.logger.Println("checking tags")
	for k, v := range p.cfg.Tags() {

		p.logger.Println("checking", k, v)
		p.Send(&messages.DoHave{k, v})

		select {
		case <-time.After(5 * time.Second):
			p.logger.Println("timeout")
			p.Remove()
			return nil
		case m, ok := <-c:
			if !ok {
				p.Remove()
				p.logger.Println("incoming closed")
				return nil
			}
			have := m.(*messages.Have)

			if !have.Have || have.TagKey != k || have.TagValue != v {
				p.Remove()
				p.logger.Printf("Tag %s:%s not supported", k, v)
				return nil
			}
		}
	}
	p.Send(&messages.Connect{})
	p.logger.Println("Connected successfully")
	p.connected.Complete(p)
	return nil
}
