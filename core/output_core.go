package core

import (
	"fmt"
	"time"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/joernweissenborn/eventual2go"
)

type OutputCore struct {
	*core
	listener map[string]map[uuid.UUID]network.Connection
	observer map[string]map[uuid.UUID]network.Connection
	requests *message.RequestStream
}

func NewOutputCore(desc descriptor.Descriptor, usrCfg *config.UserConfig,
	tracker network.Tracker, provider ...network.Provider) (o OutputCore, err error) {

	tags := desc.AsTagset()
	tags.Merge(usrCfg.Tags)

	intCfg := config.NewInternalConfig(true, tags)

	c, err := initCore(desc,
		&config.Config{
			Internal: intCfg,
			User:     usrCfg,
		}, tracker, provider...)

	o = OutputCore{
		core:     c,
		listener: map[string]map[uuid.UUID]network.Connection{},
		observer: map[string]map[uuid.UUID]network.Connection{},
	}

	o.requests = &message.RequestStream{
		Stream: o.provider.Messages().
			Where(network.OfType(message.REQUEST)).
			Transform(network.ToMessage),
	}

	o.provider.Messages().Where(network.OfType(message.HELLO)).Listen(o.onHello)

	o.Reactor.React(afterConnectedEvent{}, o.onAfterConnected)

	o.React(replyEvent{}, o.onReply)

	o.AddStream(startListenEvent{}, o.provider.Messages().Where(network.OfType(message.STARTLISTEN)).Stream)
	o.React(startListenEvent{}, o.onStartListen)

	o.AddStream(stopListenEvent{}, o.provider.Messages().Where(network.OfType(message.STOPLISTEN)).Stream)
	o.React(stopListenEvent{}, o.onStopListen)

	o.AddStream(startObserveEvent{}, o.provider.Messages().Where(network.OfType(message.STARTOBSERVE)).Stream)
	o.React(startObserveEvent{}, o.onStartObserve)

	o.AddStream(stopObserveEvent{}, o.provider.Messages().Where(network.OfType(message.STOPOBSERVE)).Stream)
	o.React(stopObserveEvent{}, o.onStopObserve)

	o.AddStream(getPropertyEvent{}, o.provider.Messages().Where(network.OfType(message.GETPROPERTY)).Stream)
	o.React(getPropertyEvent{}, o.onGetProperty)

	for p, obs := range o.properties {
		o.React(setPropertyEvent{p}, o.onSetProperty(p))
		o.AddObservable(setPropertyEvent{p}, obs)
	}

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
				o.log.Debug("Got message DOHAVE from ", conn.UUID)
				tag := nmsg.Decode().(*message.DoHave).Tag
				have = o.config.Internal.Tags.Has(tag)
				if have {
					o.log.Debugf("Tag '%s' is supported", tag)
					next = in.First()
				}
				conn.Send(
					&message.Have{
						Have: have,
						Tag:  tag,
					},
				)

				if !have {
					o.log.Debugf("Tag '%s' is not supported aborting",tag)
					conn.Close()
					return
				}
			case message.CONNECT:
				o.Fire(connectEvent{}, conn)
				o.log.Debug("Got message CONNECT from ", conn.UUID)
				return
			}
		}

	}
}

func (o *OutputCore) onAfterConnected(d eventual2go.Data) {
	conn := d.(network.Connection)
	o.log.Debug("Sending CONNECT to",conn.UUID)
	conn.Send(&message.Connect{})

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
		o.listener[function] = map[uuid.UUID]network.Connection{}
	}

	o.listener[function][m.Sender] = o.connections[m.Sender]
}

func (o *OutputCore) onStopListen(d eventual2go.Data) {
	m := d.(network.Message)

	function := m.Decode().(*message.StopListen).Function

	o.log.Infof("Listener stopped listening to function '%s': %s", function, m.Sender)

	if _, ok := o.listener[function]; ok {
		delete(o.listener[function], m.Sender)
	}
}

func (o *OutputCore) onSetProperty(property string) eventual2go.Subscriber {
	return func(d eventual2go.Data) {
		v := d.([]byte)

		o.log.Debugf("Updating property '%s'", property)

		m := &message.SetProperty{
			Name:  property,
			Value: v,
		}
		for peer, conn := range o.observer[property] {
			o.log.Debugf("Sending property update of '%s' to %s", property, peer)
			conn.Send(m)
		}
	}
}

func (o *OutputCore) onGetProperty(d eventual2go.Data) {
	m := d.(network.Message)

	property := m.Decode().(*message.GetProperty).Name

	o.log.Debugf("Got request for property '%s' from %s", property, m.Sender)

	o.sendProperty(m.Sender, property)
}

func (o *OutputCore) onStartObserve(d eventual2go.Data) {
	m := d.(network.Message)

	property := m.Decode().(*message.StartObserve).Property

	o.log.Infof("Got new observer for property '%s': %s", property, m.Sender)

	if _, ok := o.listener[property]; !ok {
		o.observer[property] = map[uuid.UUID]network.Connection{}
	}
	o.observer[property][m.Sender] = o.connections[m.Sender]
	o.sendProperty(m.Sender, property)
}

func (o *OutputCore) sendProperty(peer uuid.UUID, property string) {
	o.log.Debugf("Sending Value of property '%s' to %s", property, peer)
	o.connections[peer].Send(&message.SetProperty{
		Name:  property,
		Value: o.properties[property].Value().([]byte),
	})
}

func (o *OutputCore) onStopObserve(d eventual2go.Data) {
	m := d.(network.Message)

	property := m.Decode().(*message.StopObserve).Property

	o.log.Infof("Observer stopped observing property '%s': %s", property, m.Sender)

	if _, ok := o.observer[property]; ok {
		delete(o.listener[property], m.Sender)
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

func (o OutputCore) SetProperty(property string, value []byte) (err error) {
	obs, ok := o.properties[property]
	if !ok {
		err = fmt.Errorf("Can't set unknown property '%s'", property)
		return
	}
	obs.Change(value)
	return
}

// Emit acts like a ThingiverseIO Trigger, which is initiated by the Output.
func (o *OutputCore) Emit(function string, inparams []byte, outparams []byte) (err error) {
	if !o.descriptor.HasFunction(function) {
		err = fmt.Errorf("Function '%s' is not in descriptor", function)
		return
	}


	uuid := o.UUID()
	req := message.NewRequest(uuid, function, message.TRIGGER, inparams)
	o.Reply(req, outparams)
	return
}
