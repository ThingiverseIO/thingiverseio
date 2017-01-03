package connection

import (
	"net"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pull"

	"fmt"
	"time"

	"github.com/ThingiverseIO/thingiverseio/service/messages"
	"github.com/joernweissenborn/eventual2go"
)

type Incoming struct {
	addr   string
	port   int
	skt    mangos.Socket
	in     *messages.FlatMessageStreamController
	close  *eventual2go.Completer
	closed *eventual2go.Completer
}

func NewIncoming(addr string) (i *Incoming, err error) {
	i = &Incoming{
		addr:   addr,
		in:     messages.NewFlatMessageStreamController(),
		close:  eventual2go.NewCompleter(),
		closed: eventual2go.NewCompleter(),
	}
	err = i.setupSocket()
	if err == nil {
		go i.listen()
	}
	return
}

func (i *Incoming) In() *messages.FlatMessageStream {
	return i.in.Stream()
}

func (i *Incoming) Messages() *messages.MessageStream {
	return &messages.MessageStream{i.In().Transform(ToMessage)}
}

func (i *Incoming) Addr() (addr string) {
	return i.addr
}

func (i *Incoming) Port() (port int) {
	return i.port
}

func (i *Incoming) setupSocket() (err error) {
	i.port = getRandomPort()
	i.skt, err = pull.NewSocket()
	if err != nil {
		return
	}
	i.skt.SetOption("RECV-DEADLINE", 100*time.Millisecond)

	err = i.skt.Listen(fmt.Sprintf("tcp://%s:%d", i.addr, i.port))
	return
}

func getRandomPort() int {
	l, err := net.Listen("tcp", ":0") // listen on address
	if err != nil {
		panic(fmt.Sprintf("Could not find a free port %v", err))
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func (i *Incoming) listen() {

	for {
		if i.close.Completed() {
			err := i.skt.Close()
			i.closed.Complete(err)
			return
		}
		msg, err := i.skt.Recv()
		if err != nil {
			continue
		}
		if m, ok := messages.Decode(msg); ok {
			i.in.Add(m)
		}
	}
}

func (i *Incoming) Shutdown(eventual2go.Data) (err error) {
	i.close.Complete(nil)
	i.closed.Future().WaitUntilComplete()
	err, _ = i.closed.Future().GetResult().(error)
	return
}
