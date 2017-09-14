package beacon

import (
	"errors"
	"net"
	"time"

	"github.com/joernweissenborn/eventual2go"
)

type listener struct {
	buffer  []byte
	signals *SignalStreamController
	socket  *net.UDPConn
	stop    *eventual2go.Completer
}

func newListener(port int) (l *listener, err error) {
	ip := net.IPv4(224, 0, 0, 165)

	ifis, err := net.Interfaces()
	if err != nil {
		return
	}

	//get the first multicast adapter for listening
	var ifi *net.Interface
	for _, i := range ifis {
		if i.Flags&net.FlagMulticast == net.FlagMulticast {
			ifi = &i
			break
		}
	}

	if ifi == nil {
		err = errors.New("Could not find multicast adapter")
		return
	}

	socket, err := net.ListenMulticastUDP(
		"udp4",
		ifi,
		&net.UDPAddr{
			IP:   ip,
			Port: port,
		},
	)
	if err != nil {
		return
	}

	l = &listener{
		buffer:  make([]byte, 1024),
		signals: NewSignalStreamController(),
		socket:  socket,
		stop:    eventual2go.NewCompleter(),
	}

	return
}

func (l listener) listen() {
	stop := l.stop.Future()

	for !stop.Completed() {
		l.read()
	}

	l.socket.Close()
}

func (l listener) read() {

	l.socket.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	read, remoteAddr, err := l.socket.ReadFromUDP(l.buffer)
	if err == nil {
		data := make([]byte, read)
		copy(data, l.buffer)
		l.signals.Add(Signal{remoteAddr.IP[len(remoteAddr.IP)-4:], data})
	}
}

func (l listener) close() {
	l.stop.Complete(nil)
}
