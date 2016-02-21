package peer

import (
	"fmt"
	"log"
	"time"

	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/joernweissenborn/thingiverse.io/service"
	"github.com/joernweissenborn/thingiverse.io/service/connection"
	"github.com/joernweissenborn/thingiverse.io/service/messages"
)

type Peer struct {
	uuid string
	cfg  *config.Config

	incoming *connection.Incoming

	initialized *eventual2go.Completer

	msgIn  *messages.MessageStream
	msgOut *connection.Outgoing

	removed *eventual2go.Completer

	logger *log.Logger
}

func New(uuid, address string, port int, incoming *connection.Incoming, cfg *config.Config) (p *Peer, err error) {
	p = &Peer{
		uuid:     uuid,
		incoming: incoming,
		msgIn:    &messages.MessageStream{incoming.In().Where(connection.IsMsgFromSender(uuid)).Where(validMsg).Transform(transformToMessage)},
		removed:  eventual2go.NewCompleter(),
		logger:   log.New(cfg.Logger(), fmt.Sprintf("PEER %s@%s ", uuid[:5], cfg.UUID()[:5]), 0),
	}

	p.msgIn.Where(messages.Is(messages.HELLO_OK)).Listen(p.onHelloOk)
	p.msgOut, err = connection.NewOutgoing(cfg.UUID(), address, port)
	if err != nil {
		return
	}
	p.removed.Future().Then(p.closeOutgoing)
	return
}

func NewFromHello(uuid string, m *messages.Hello, incoming *connection.Incoming, cfg *config.Config) (p *Peer, err error) {
	p, err = New(uuid, m.Address, m.Port, incoming, cfg)

	p.logger.Println("Received HELLO")
	if err != nil {
		return
	}

	p.Send(&messages.HelloOk{})

	p.initialized = eventual2go.NewTimeoutCompleter(5 * time.Second)
	p.initialized.Future().Err(p.onTimeout)

	return
}

func (p *Peer) InitConnection() {

	if p.initialized != nil {
		return
	}

	hello := &messages.Hello{p.incoming.Addr(), p.incoming.Port()}

	p.Send(hello)

	p.initialized = eventual2go.NewTimeoutCompleter(5 * time.Second)
	p.initialized.Future().Err(p.onTimeout)

}

func (p *Peer) Messages() *messages.MessageStream {
	return p.msgIn
}

func (p *Peer) onTimeout(error) (eventual2go.Data, error) {
	p.removed.Complete(nil)
	return nil, nil
}

func (p *Peer) onHelloOk(messages.Message) {
	p.logger.Println("Received HELLO_OK")
	if !p.initialized.Completed() {
		p.initialized.Complete(nil)
		p.Send(&messages.HelloOk{})
	}
}

func (p *Peer) Removed() *eventual2go.Future {
	return p.removed.Future()
}

func (p *Peer) Send(m messages.Message) {
	p.msgOut.Send(messages.Flatten(m))
}

func (p *Peer) closeOutgoing(eventual2go.Data) eventual2go.Data {
	p.msgOut.Close()
	return nil
}

func validMsg(m connection.Message) bool {
	if len(m.Payload) < 2 {
		return false
	}
	p := []byte(m.Payload[0])[0]

	if p != service.PROTOCOLL_SIGNATURE {
		return false
	}
	return true
}

func transformToMessage(d eventual2go.Data) eventual2go.Data {
	m := d.(connection.Message)
	return messages.Unflatten(m.Payload)
}
