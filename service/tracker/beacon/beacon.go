package beacon

import (
	"bytes"
	"errors"
	"log"
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

	logger *log.Logger

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

	b.init()

	err = b.setup()

	return
}

func (b *Beacon) Run() {
	b.logger.Println("Running")
	go b.listen()
}

func (b *Beacon) init() {
	b.conf.init()
	b.logger = log.New(b.conf.Logger, "BEACON ", 0)
}

func (b *Beacon) setup() (err error) {
	b.logger.Println("Setting up")
	err = b.setupListener()
	if err != nil {
		return
	}
	err = b.setupSender()
	return
}

func (b *Beacon) Stop() {
	b.logger.Println("Stopping")
	b.Silence()
	b.stop.Complete(nil)
	b.logger.Println("Stopped")
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

	b.logger.Printf("Multicast Address is %s", ip)

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
			b.logger.Println("Stopped to listen")
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
	if !b.signals.Closed().Completed() && err == nil {
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
		b.logger.Println("Silencing")
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
	b.logger.Println("Started to ping")

	t := time.NewTimer(b.conf.PingInterval)
	silence := b.silence.Future().AsChan()
	for {
		select {
		case <-silence:
			b.logger.Println("Silenced")
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
