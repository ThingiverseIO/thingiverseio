package beacon

import (
	"bytes"
	"time"

	"github.com/joernweissenborn/eventual2go"
)

type Beacon struct {
	conf     *Config
	listener *listener
	sender   *sender
	silence  *eventual2go.Completer
}

func New(conf *Config) (b *Beacon, err error) {

	listener, err := newListener(conf.Address, conf.Port)
	if err != nil {
		return
	}

	sender, err := newSender(conf.Address, conf.Port)
	if err != nil {
		return
	}

	b = &Beacon{
		conf:     conf,
		listener: listener,
		sender:   sender,
		silence:  eventual2go.NewCompleter(),
	}

	return
}

func (b *Beacon) Run() {
	go b.listener.listen()
}

func (b *Beacon) Stop() {
	b.Silence()
	b.listener.close()
	b.sender.close()
}

func (b *Beacon) Signals() *SignalStream {
	return b.listener.signals.Stream().Where(b.noEcho)
}
func (b *Beacon) Silence() {
	if !b.silence.Completed() {
		b.silence.Complete(nil)
	}
}

func (b *Beacon) Silent() bool {
	return b.silence.Completed()
}

func (b *Beacon) Ping() {
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
			b.sender.send(b.conf.Payload)
			t.Reset(b.conf.PingInterval)
		}
	}
}

func (b *Beacon) noEcho(d Signal) bool {
	return !bytes.Equal(d.Data, b.conf.Payload)
}
