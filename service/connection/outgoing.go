package connection

import (
	"fmt"
	"sync"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/push"
	"github.com/joernweissenborn/eventual2go"
)

type Outgoing struct {
	m      *sync.Mutex
	skt    mangos.Socket
	closed *eventual2go.Completer
	uuid   config.UUID
}

func NewOutgoing(uuid config.UUID, targetAddress string, targetPort int) (out *Outgoing, err error) {

	skt, err := push.NewSocket()
	if err != nil {
		return
	}

	err = skt.Dial(fmt.Sprintf("tcp://%s:%d", targetAddress, targetPort))
	if err != nil {
		return
	}

	out = &Outgoing{
		m:      &sync.Mutex{},
		skt:    skt,
		closed: eventual2go.NewCompleter(),
		uuid:   uuid,
	}

	return
}

func (o *Outgoing) Send(data []byte) error {
	return o.send(data)
}

func (o *Outgoing) Close() {
	o.m.Lock()
	defer o.m.Unlock()

	if err := o.skt.Close(); err != nil {
		o.closed.CompleteError(err)
		return
	}
	o.closed.Complete(nil)
}

func (o *Outgoing) send(d []byte) (err error) {
	o.m.Lock()
	defer o.m.Unlock()
	err = o.skt.Send(d)
	if err != nil {
		if !o.closed.Completed() {
			o.Close()
		}
	}
	return
}
