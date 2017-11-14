package core

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/uuid"
	"github.com/joernweissenborn/eventual2go"
)

type OutputCore struct {
	*core
	listener map[string]map[uuid.UUID]network.Connection
	observer map[string]map[uuid.UUID]network.Connection
	consumer map[string]map[uuid.UUID]network.Connection
	requests *message.RequestStream
}

func NewOutputCore(desc descriptor.Descriptor, usrCfg *config.UserConfig,
	tracker network.Tracker, transport ...network.Transport) (o OutputCore, err error) {

	tags := desc.AsTagset()
	tags.Merge(usrCfg.Tags)

	intCfg := config.NewInternalConfig(true, tags)

	c, err := initCore(desc,
		&config.Config{
			Internal: intCfg,
			User:     usrCfg,
		}, tracker, transport...)
	if err != nil {
		return
	}

	o = OutputCore{
		core:     c,
		listener: map[string]map[uuid.UUID]network.Connection{},
		observer: map[string]map[uuid.UUID]network.Connection{},
		consumer: map[string]map[uuid.UUID]network.Connection{},
	}

	o.requests = &message.RequestStream{
		Stream: o.transport.Packages().
			TransformWhere(network.ToMessage, network.OfType(message.REQUEST)),
	}

	o.transport.Packages().Where(network.OfType(message.HELLO)).ListenNonBlocking(o.onHello)

	o.r.React(afterConnectedEvent{}, o.onAfterConnected)

	o.r.React(replyEvent{}, o.onReply)

	o.r.AddStream(startListenEvent{}, o.transport.Packages().Where(network.OfType(message.STARTLISTEN)).Stream)
	o.r.React(startListenEvent{}, o.onStartListen)

	o.r.AddStream(stopListenEvent{}, o.transport.Packages().Where(network.OfType(message.STOPLISTEN)).Stream)
	o.r.React(stopListenEvent{}, o.onStopListen)

	o.r.AddStream(startConsumeEvent{}, o.transport.Packages().Where(network.OfType(message.STARTCONSUME)).Stream)
	o.r.React(startConsumeEvent{}, o.onStartConsume)

	o.r.AddStream(stopConsumeEvent{}, o.transport.Packages().Where(network.OfType(message.STOPCONSUME)).Stream)
	o.r.React(stopConsumeEvent{}, o.onStopConsume)

	o.r.AddStream(startObserveEvent{}, o.transport.Packages().Where(network.OfType(message.STARTOBSERVE)).Stream)
	o.r.React(startObserveEvent{}, o.onStartObserve)

	o.r.AddStream(stopObserveEvent{}, o.transport.Packages().Where(network.OfType(message.STOPOBSERVE)).Stream)
	o.r.React(stopObserveEvent{}, o.onStopObserve)

	o.r.AddStream(getPropertyEvent{}, o.transport.Packages().Where(network.OfType(message.GETPROPERTY)).Stream)
	o.r.React(getPropertyEvent{}, o.onGetProperty)

	for p, obs := range o.properties {
		o.r.React(setPropertyEvent{p}, o.onSetProperty(p))
		o.r.AddObservable(setPropertyEvent{p}, obs)
	}

	for s, stream := range o.streams {
		o.r.React(addStreamEvent{s}, o.onAddStream(s))
		o.r.AddStream(addStreamEvent{s}, stream.Stream())
	}

	return

}

func (o OutputCore) onHello(p network.Package) {

	o.log.Debugf("Received HELLO message from %s", p.Sender)

	msg := p.Decode().(*message.Hello)
	o.log.Debug("Peer Details:\n", hex.Dump(msg.NetworkDetails[0]))

	conn, err := o.transport.Connect(msg.NetworkDetails, p.Sender)
	if err != nil {
		o.log.Errorf("Error connecting to %s: %s", p.Sender, err)
		return
	}

	in := o.transport.Packages().Where(network.FromSender(p.Sender))
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
					o.log.Debugf("Tag '%s' is not supported aborting", tag)
					conn.Close()
					return
				}
			case message.CONNECT:
				o.r.Fire(connectEvent{}, conn)
				o.log.Debug("Got message CONNECT from ", conn.UUID)
				return
			}
		}

	}
}

func (o *OutputCore) onAfterConnected(d eventual2go.Data) {
	conn := d.(network.Connection)
	o.log.Debug("Sending CONNECT to", conn.UUID)
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
				if conn, ok := o.connections[uuid]; ok {
					conn.Send(result)
				}
				// if something is wrong and client
			}
		}
	}
}
func (o *OutputCore) onStartListen(d eventual2go.Data) {
	p := d.(network.Package)

	function := p.Decode().(*message.StartListen).Function

	o.log.Infof("Got new listener for function '%s': %s", function, p.Sender)

	if _, ok := o.listener[function]; !ok {
		o.listener[function] = map[uuid.UUID]network.Connection{}
	}

	o.listener[function][p.Sender] = o.connections[p.Sender]
}

func (o *OutputCore) onStopListen(d eventual2go.Data) {
	p := d.(network.Package)

	function := p.Decode().(*message.StopListen).Function

	o.log.Infof("Listener stopped listening to function '%s': %s", function, p.Sender)

	if _, ok := o.listener[function]; ok {
		delete(o.listener[function], p.Sender)
	}
}

func (o *OutputCore) onAddStream(stream string) eventual2go.Subscriber {
	return func(d eventual2go.Data) {
		v := d.([]byte)

		o.log.Debugf("Adding to stream '%s'", stream)

		m := &message.AddStream{
			Name:  stream,
			Value: v,
		}
		for peer, conn := range o.consumer[stream] {
			o.log.Debugf("Sending value on stream  '%s' to %s", stream, peer)
			conn.Send(m)
		}
	}
}

func (o *OutputCore) onStartConsume(d eventual2go.Data) {
	p := d.(network.Package)

	stream := p.Decode().(*message.StartConsume).Stream

	o.log.Infof("Got new consumer for stream '%s': %s", stream, p.Sender)

	if _, ok := o.listener[stream]; !ok {
		o.consumer[stream] = map[uuid.UUID]network.Connection{}
	}

	o.consumer[stream][p.Sender] = o.connections[p.Sender]
}

func (o *OutputCore) onStopConsume(d eventual2go.Data) {
	p := d.(network.Package)

	stream := p.Decode().(*message.StopConsume).Stream

	o.log.Infof("Consumer stopped consuming stream '%s': %s", stream, p.Sender)

	if _, ok := o.consumer[stream]; ok {
		delete(o.consumer[stream], p.Sender)
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
	p := d.(network.Package)

	property := p.Decode().(*message.GetProperty).Name

	o.log.Debugf("Got request for property '%s' from %s", property, p.Sender)

	o.sendProperty(p.Sender, property)
}

func (o *OutputCore) onStartObserve(d eventual2go.Data) {
	p := d.(network.Package)

	property := p.Decode().(*message.StartObserve).Property

	o.log.Infof("Got new observer for property '%s': %s", property, p.Sender)

	if _, ok := o.observer[property]; !ok {
		o.observer[property] = map[uuid.UUID]network.Connection{}
	}
	o.observer[property][p.Sender] = o.connections[p.Sender]
	o.sendProperty(p.Sender, property)
}

func (o *OutputCore) sendProperty(peer uuid.UUID, property string) {
	o.log.Debugf("Sending Value of property '%s' to %s", property, peer)
	o.connections[peer].Send(&message.SetProperty{
		Name:  property,
		Value: o.properties[property].Value().([]byte),
	})
}

func (o *OutputCore) onStopObserve(d eventual2go.Data) {
	p := d.(network.Package)

	property := p.Decode().(*message.StopObserve).Property

	o.log.Infof("Observer stopped observing property '%s': %s", property, p.Sender)

	if _, ok := o.observer[property]; ok {
		delete(o.observer[property], p.Sender)
	}
}

// Reply reponds the given output parameter to all interested Inputs of a given request.
func (o OutputCore) Reply(r *message.Request, params []byte) (err error) {
	res := message.NewResult(o.UUID(), r, params)
	o.r.Fire(replyEvent{}, res)
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

func (o OutputCore) AddStream(stream string, value []byte) (err error) {
	s, ok := o.streams[stream]
	if !ok {
		err = fmt.Errorf("Can't add to unknown stream '%s'", stream)
		return
	}
	s.Add(value)
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
