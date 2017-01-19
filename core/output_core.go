package core

import (
	"time"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/network"
)

type OutputCore struct {
	Core
}

func NewOutputCore(desc descriptor.Descriptor, usrCfg *config.UserConfig,
	tracker network.Tracker, provider ...network.Provider) (o OutputCore, err error) {

	tags := desc.AsTagset()
	tags.Merge(usrCfg.Tags)

	intCfg := config.NewInternalConfig(true, tags)

	c, err := Initialize(&config.Config{
		Internal: intCfg,
		User:     usrCfg,
	}, tracker, provider...)

	o = OutputCore{
		Core: c,
	}

	o.provider.Messages().Where(network.OfType(message.HELLO)).Listen(o.onHello)

	return

}

func (o OutputCore) onHello(m network.Message) {

	o.log.Debugf("Received HELLO message from %s", m.Sender)

	msg := m.Decode().(*message.Hello)

	conn, err := o.provider.Connect(msg.NetworkDetails, m.Sender)
	if err != nil {
		return
	}

	in := o.provider.Messages().Where(network.FromSender(m.Sender))
	defer in.Close()

	next := in.First()

	have := o.config.Internal.Tags.Has(msg.Tag)
	conn.Send(&message.HelloOk{Have: have})

	if !have {
		o.log.Debug("Peer is not supported, aborting.")
		conn.Close()
		return
	}

	for {
		if next.WaitUntilTimeout(1 * time.Second) {

			switch nmsg := next.GetResult(); nmsg.Type {
			case message.DOHAVE:
				o.log.Debug("Got message DOHAVE")
				tag := nmsg.Decode().(*message.DoHave).Tag
				have = o.config.Internal.Tags.Has(tag)
				if have {
					next = in.First()
				}
				conn.Send(&message.Have{
					Have: have,
					Tag:  tag,
				})

				if !have {
					conn.Close()
					return
				}
			case message.CONNECT:
				o.log.Debug("Got message CONNECT")
				o.Fire(connectEvent{}, conn)
				return
			}
		}

	}
}
