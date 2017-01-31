package core

import (
	"time"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/joernweissenborn/eventual2go"
)

type OutputCore struct {
	*Core
	listener map[string][]uuid.UUID
	requests *message.RequestStream
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

	o.requests = &message.RequestStream{
		Stream: o.provider.Messages().
			Where(network.OfType(message.REQUEST)).
			Transform(network.ToMessage),
	}

	o.provider.Messages().Where(network.OfType(message.HELLO)).Listen(o.onHello)

	o.React(replyEvent{}, o.onReply)

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

			switch nmsg := next.Result(); nmsg.Type {
			case message.DOHAVE:
				o.log.Debug("Got message DOHAVE")
				tag := nmsg.Decode().(*message.DoHave).Tag
				have = o.config.Internal.Tags.Has(tag)
				if have {
					next = in.First()
				}
				conn.Send(
					&message.Have{
						Have: have,
						Tag:  tag,
					},
				)

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

func (o OutputCore) onReply(d eventual2go.Data) {
	result := d.(*message.Result)
	o.log.Debugf("Replying to function '%s', calltype is %s", result.Request.Function, result.Request.CallType)

	switch result.Request.CallType {
	case message.CALL, message.CALLALL:
		if conn, ok := o.connections[result.Request.Input]; ok {
			o.log.Debug("Delivering to", result.Request.Input)
			conn.Send(result)
		} else {
			o.log.Debug("Abborting delivery, peer does not exist anymore", result.Request.Input)
		}

	case message.TRIGGER, message.TRIGGERALL:
		if ls, ok := o.listener[result.Request.Function]; ok {
			for _, uuid := range ls {
				o.log.Debug("Delivering to", uuid)
				o.connections[uuid].Send(result)
			}
		}
	}
}

// Reply reponds the given output parameter to all interested Inputs of a given request.
func (o OutputCore) Reply(r *message.Request, params []byte) (err error) {
	res := message.NewResult(o.UUID(), r, params)
	o.Fire(replyEvent{}, res)
	return
}

func (o OutputCore) RequestStream() *message.RequestStream {
	return o.requests
}
