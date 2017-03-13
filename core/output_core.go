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
	listener map[string]map[uuid.UUID]interface{}
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
		Core:     c,
		listener: map[string]map[uuid.UUID]interface{}{},
	}

	o.requests = &message.RequestStream{
		Stream: o.provider.Messages().
			Where(network.OfType(message.REQUEST)).
			Transform(network.ToMessage),
	}

	o.provider.Messages().Where(network.OfType(message.HELLO)).Listen(o.onHello)

	o.React(replyEvent{}, o.onReply)

	o.AddStream(startListenEvent{}, o.provider.Messages().Where(network.OfType(message.STARTLISTEN)).Stream)
	o.React(startListenEvent{}, o.onStartListen)

	o.AddStream(stopListenEvent{}, o.provider.Messages().Where(network.OfType(message.STOPLISTEN)).Stream)
	o.React(stopListenEvent{}, o.onStopListen)

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
				o.Fire(connectEvent{}, conn)
				o.log.Debug("Got message CONNECT")
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
			o.log.Debugf("Delivering reply for %s to %s", result.Request.UUID, result.Request.Input)
			conn.Send(result)
		} else {
			o.log.Debug("Aborting delivery, peer does not exist anymore", result.Request.Input)
		}

	case message.TRIGGER, message.TRIGGERALL:
		if ls, ok := o.listener[result.Request.Function]; ok {
			for uuid := range ls {
				o.log.Debug("Delivering to", uuid)
				o.connections[uuid].Send(result)
			}
		}
	}
}

func (o *OutputCore) onStartListen(d eventual2go.Data) {
	m := d.(network.Message)

	function := m.Decode().(*message.StartListen).Function

	o.log.Infof("Got new listener for function '%s': %s", function, m.Sender)

	if _, ok := o.listener[function]; !ok {
		o.listener[function] = map[uuid.UUID]interface{}{}
	}

	o.listener[function][m.Sender] = nil
}

func (o *OutputCore) onStopListen(d eventual2go.Data) {
	m := d.(network.Message)

	function := m.Decode().(*message.StopListen).Function

	o.log.Infof("Listener stopped listening to function '%s': %s", function, m.Sender)

	if _, ok := o.listener[function]; ok {
		delete(o.listener[function], m.Sender)
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
