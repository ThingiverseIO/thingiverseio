package connection

import (
	"net"

	"github.com/pebbe/zmq4"

	"fmt"
	"time"

	"github.com/joernweissenborn/eventual2go"
)

type Incoming struct {
	addr string
	port int
	skt  *zmq4.Socket
	in   *eventual2go.StreamController
	stop *eventual2go.Completer
}

func NewIncoming(addr string) (i *Incoming, err error) {
	i = &Incoming{
		addr: addr,
		in:   eventual2go.NewStreamController(),
		stop: eventual2go.NewCompleter(),
	}
	err = i.setupSocket()
	if err == nil {
		go i.listen()
	}
	return
}

func (i *Incoming) In() *eventual2go.Stream {
	return i.in.Stream
}

func (i *Incoming) Addr() (addr string) {
	return i.addr
}

func (i *Incoming) Port() (port int) {
	return i.port
}

func (i *Incoming) setupSocket() (err error) {
	i.port = getRandomPort()
	i.skt, err = zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		return
	}
	err = i.skt.Bind(fmt.Sprintf("tcp://%s:%d", i.addr, i.port))
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
	poller := zmq4.NewPoller()
	poller.Add(i.skt, zmq4.POLLIN)

	for {
		if i.stop.Completed() {
			i.skt.Close()
			return
		}
		sockets, err := poller.Poll(100 * time.Millisecond)
		if err != nil {
			continue
		}
		for range sockets {
			msg, err := i.skt.RecvMessage(0)
			if err == nil {
				i.in.Add(Message{i.addr, msg[0], msg[1:]})
			}
		}
	}
}

func (i *Incoming) Close() {
	i.stop.Complete(nil)
}