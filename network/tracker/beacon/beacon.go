package beacon

import (
	"bytes"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/joernweissenborn/eventual2go"
)

type Beacon struct {
	m    *sync.Mutex
	conf *Config

	incoming *net.UDPConn
	outgoing *net.UDPConn

	signals *SignalStreamController

	silence *eventual2go.Completer

	stop *eventual2go.Completer
}

func New(conf *Config) (b *Beacon, err error) {
	b = &Beacon{
		m:       &sync.Mutex{},
		conf:    conf,
		signals: NewSignalStreamController(),
		silence: eventual2go.NewCompleter(),
		stop:    eventual2go.NewCompleter(),
	}

	err = b.setup()

	return
}

func (b *Beacon) Run() {
	go b.listen()
}

func (b *Beacon) setup() (err error) {
	if err = b.setupListener(); err != nil {
		return
	}
	err = b.setupSender()
	return
}

func (b *Beacon) Stop() {
	b.Silence()
	b.stop.Complete(nil)
}

func (b *Beacon) setupSender() (err error) {

	b.outgoing, err = net.DialUDP(
		"udp4",
		&net.UDPAddr{
			IP:   net.ParseIP(b.conf.Addr),
			Port: 0},
		&net.UDPAddr{
			IP:   net.IPv4bcast,
			Port: b.conf.Port,
		},
	)
	if err == nil {
		b.stop.Future().Then(b.closeListener)
	}
	return
}

func (b *Beacon) closeListener(eventual2go.Data) eventual2go.Data {
	b.m.Lock()
	defer b.m.Unlock()
	b.incoming.Close()
	return nil
}

func (b *Beacon) setupListener() (err error) {
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

	b.incoming, err = net.ListenMulticastUDP(
		"udp4",
		ifi,
		&net.UDPAddr{
			IP:   ip,
			Port: b.conf.Port,
		},
	)

	if err == nil {
		b.stop.Future().Then(b.closeOutgoing)
	}
	return
}

func (b *Beacon) closeOutgoing(eventual2go.Data) eventual2go.Data {
	b.m.Lock()
	defer b.m.Unlock()
	b.outgoing.Close()
	return nil
}

func (b *Beacon) listen() {

	stop := b.stop.Future().AsChan()
	c := make(chan struct{})
	go b.getSignal(c)
	for {
		select {
		case <-stop:
			return

		case <-c:
			go b.getSignal(c)
		}
	}

}

func (b *Beacon) getSignal(c chan struct{}) {
	b.m.Lock()
	defer b.m.Unlock()
	data := make([]byte, 1024)
	b.incoming.SetReadDeadline(time.Now().Add(1 * time.Second))
	read, remoteAddr, err := b.incoming.ReadFromUDP(data)
	if err == nil {
		b.signals.Add(Signal{remoteAddr.IP[len(remoteAddr.IP)-4:], data[:read]})
	}
	c <- struct{}{}
}

func (b *Beacon) Signals() *SignalStream {
	return b.signals.Stream().Where(b.noEcho)
}
func (b *Beacon) Silence() {
	b.m.Lock()
	defer b.m.Unlock()
	if !b.silence.Completed() {
		b.silence.Complete(nil)
	}
}

func (b *Beacon) Silent() bool {
	return b.silence.Completed()
}

func (b *Beacon) Ping() {
	b.m.Lock()
	defer b.m.Unlock()
	if b.silence.Completed() {
		b.silence = eventual2go.NewCompleter()
	}
	go b.ping()
}

func (b *Beacon) ping() {
	t := time.NewTimer(b.conf.PingInterval)
	silence := b.silence.Future().AsChan()
	for {
		select {
		case <-silence:
			return

		case <-t.C:
			b.outgoing.Write(b.conf.Payload)
			t.Reset(b.conf.PingInterval)
		}
	}

}

func (b *Beacon) noEcho(d Signal) bool {
	return !bytes.Equal(d.Data, b.conf.Payload)
}
