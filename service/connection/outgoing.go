package connection

import (
	"fmt"

	"github.com/joernweissenborn/eventual2go"
	"github.com/pebbe/zmq4"
)

type Outgoing struct {
	skt    *zmq4.Socket
	out    *eventual2go.StreamController
	closed *eventual2go.Completer
}

func NewOutgoing(uuid string, targetAddress string, targetPort int) (out *Outgoing, err error) {

	skt, err := zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		return
	}

	err = skt.SetIdentity(uuid)
	if err != nil {
		return
	}

	err = skt.Connect(fmt.Sprintf("tcp://%s:%d", targetAddress, targetPort))
	if err != nil {
		return
	}

	out = &Outgoing{
		skt:    skt,
		out:    eventual2go.NewStreamController(),
		closed: eventual2go.NewCompleter(),
	}

	out.out.Stream.Listen(out.send)
	out.closed.Future().Then(out.close)

	return
}

func (o *Outgoing) Sent(data [][]byte) {
	o.out.Add(data)
}

func (o *Outgoing) Close() {
	o.closed.Complete(nil)
}

func (o *Outgoing) send(d eventual2go.Data) {

	_, err := o.skt.SendMessage(d)

	if err != nil {
		o.closed.CompleteError(err)
		o.close(nil)
	}

	return
}
func (o *Outgoing) close(eventual2go.Data) eventual2go.Data {
	o.out.Close()
	return o.skt.Close()

}
