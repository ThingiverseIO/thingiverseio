package connection

import (
	"fmt"
	"sync"

	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/thingiverseio/config"
	"github.com/pebbe/zmq4"
)

type Outgoing struct {
	m      *sync.Mutex
	skt    *zmq4.Socket
	closed *eventual2go.Completer
}

func NewOutgoing(uuid config.UUID, targetAddress string, targetPort int) (out *Outgoing, err error) {

	skt, err := zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		return
	}

	err = skt.SetIdentity(string(uuid))
	if err != nil {
		return
	}

	err = skt.Connect(fmt.Sprintf("tcp://%s:%d", targetAddress, targetPort))
	if err != nil {
		return
	}

	out = &Outgoing{
		m:      &sync.Mutex{},
		skt:    skt,
		closed: eventual2go.NewCompleter(),
	}

	return
}

func (o *Outgoing) Send(data [][]byte) error {
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

func (o *Outgoing) send(d [][]byte) (err error) {
	o.m.Lock()
	defer o.m.Unlock()
	_, err = o.skt.SendMessage(d)
	if err != nil {
		if !o.closed.Completed() {
			o.Close()
		}
	}
	return
}
