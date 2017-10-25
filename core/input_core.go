package core

import (
	"encoding/hex"
	"fmt"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/joernweissenborn/eventual2go"
)

func isListenResult(r *message.Result) bool {
	return r.Request.CallType == message.TRIGGER || r.Request.CallType == message.TRIGGERALL
}

type InputCore struct {
	*core
	listenedFunctions  map[string]interface{}
	observedProperties map[string]interface{}
	consumedStreams    map[string]interface{}
	results            *message.ResultStream
}

func NewInputCore(desc descriptor.Descriptor, usrCfg *config.UserConfig,
	tracker network.Tracker, provider ...network.Provider) (i InputCore, err error) {

	tags := desc.AsTagset()
	tags.Merge(usrCfg.Tags)

	intCfg := config.NewInternalConfig(false, tags)

	c, err := initCore(desc,
		&config.Config{
			Internal: intCfg,
			User:     usrCfg,
		}, tracker, provider...)

	i = InputCore{
		core:               c,
		listenedFunctions:  map[string]interface{}{},
		observedProperties: map[string]interface{}{},
		consumedStreams:    map[string]interface{}{},
	}

	i.results = &message.ResultStream{
		Stream: i.provider.Messages().
			Where(network.OfType(message.RESULT)).
			Transform(network.ToMessage),
	}

	i.tracker.Arrivals().Where(i.isInterestingArrival).ListenNonBlocking(i.onArrival)

	i.r.React(afterConnectedEvent{}, i.onAfterConnected)

	i.r.React(startListenEvent{}, i.onStartListen)
	i.r.React(stopListenEvent{}, i.onStopListen)

	i.r.React(startConsumeEvent{}, i.onStartConsume)
	i.r.React(stopConsumeEvent{}, i.onStopConsume)

	i.r.AddStream(addStreamEvent{}, i.provider.Messages().TransformWhere(network.ToMessage, network.OfType(message.ADDSTREAM)))
	i.r.React(addStreamEvent{}, i.onAddStream)

	i.r.React(startObserveEvent{}, i.onStartObserve)
	i.r.React(stopObserveEvent{}, i.onStopObserve)

	i.r.AddStream(setPropertyEvent{}, i.provider.Messages().TransformWhere(network.ToMessage, network.OfType(message.SETPROPERTY)))
	i.r.React(setPropertyEvent{}, i.onSetProperty)

	return

}

func (i InputCore) isInterestingArrival(a network.Arrival) (is bool) {
	i.log.Debugf("Checking if arrived peer is interesting: %s", a.UUID)

	if is = a.IsOutput; !is {
		i.log.Debug("Peer is not an output.")
		return
	}
	if is = a.Supported(i.provider.Details); !is {
		i.log.Debug("Peer network is not supported")
	}
	i.log.Debug("Peer is interesting")
	return
}

func (i *InputCore) onAfterConnected(d eventual2go.Data) {
	conn := d.(network.Connection)

	for function := range i.listenedFunctions {
		i.log.Debugf("Informing %s about function listenig to '%s'", conn.UUID, function)
		conn.Send(&message.StartListen{
			Function: function,
		})
	}

	for property := range i.observedProperties {
		i.log.Debugf("Informing %s about observing property '%s'", conn.UUID, property)
		conn.Send(&message.StartObserve{
			Property: property,
		})
	}

	for stream := range i.consumedStreams {
		i.log.Debugf("Informing %s about consuming stream '%s'", conn.UUID, stream)
		conn.Send(&message.StartConsume{
			Stream: stream,
		})
	}
}

func (i InputCore) onArrival(a network.Arrival) {
	i.log.Debug("Peer arrived: ", a.UUID)
	i.log.Debug("Peer Details:\n", hex.Dump(a.Details[0]))
	conn, err := i.provider.Connect(a.Details, a.UUID)
	if err != nil {
		i.log.Error("Error connecting to peer ", conn.UUID, err)
		return
	}

	in := i.provider.Messages().Where(network.FromSender(a.UUID))
	defer in.Close()

	next := in.FirstWhere(network.OfType(message.HELLOOK))

	// Send Hello

	conn.Send(&message.Hello{
		NetworkDetails: i.provider.EncodedDetails,
		Tag:            i.config.Internal.Tags.GetFirst(),
	})

	if next.WaitUntilTimeout(connectionTimeout) {
		i.log.Debug("Received HELLOOK from ", conn.UUID)
		if next.Result().Decode().(*message.HelloOk).Have {
			i.log.Debugf("First Tag is supported by %s, checking all tags", conn.UUID)
			for _, tag := range i.config.Internal.Tags.AsArray() {

				i.log.Debugf("Checking peer %s for tag '%s'", conn.UUID, tag)
				next = in.FirstWhere(network.OfType(message.HAVE))

				// Send DoHave
				conn.Send(&message.DoHave{
					Tag: tag,
				})

				if next.WaitUntilTimeout(connectionTimeout) {
					i.log.Debug("Got message HAVE")
					if !next.Result().Decode().(*message.Have).Have {
						// Send End

						i.log.Debugf("Peer %s does not support tag '%s', aborting", conn.UUID, tag)
						conn.Send(&message.End{})
						conn.Close()
						return
					}
				} else {
					conn.Close()
					i.log.Debug("Timeout", conn.UUID)
					return
				}
			}

			// Send Connect
			conn.Send(&message.Connect{})
			// Await Confirmation
			next = in.FirstWhere(network.OfType(message.CONNECT))
			if next.WaitUntilTimeout(connectionTimeout) {
				i.log.Debug("Received CONNECT from", conn.UUID)
				i.r.Fire(connectEvent{}, conn)
				return
			}
		} else {
			i.log.Debugf("Peer %s does not support first tag , aborting", conn.UUID)
			conn.Send(&message.End{})
			conn.Close()
			return
		}
	}
	i.log.Debug("Timeout", conn.UUID)

	conn.Close()

}

func (i *InputCore) onStartListen(d eventual2go.Data) {
	function := d.(string)
	i.log.Infof("Starting to listen to function '%s'", function)
	i.listenedFunctions[function] = nil
	i.sendToAll(&message.StartListen{
		Function: function,
	})
}

func (i *InputCore) onStopListen(d eventual2go.Data) {
	function := d.(string)
	i.log.Infof("Stopping to listen to function '%s'", function)

	if _, ok := i.listenedFunctions[function]; ok {
		delete(i.listenedFunctions, function)
		i.sendToAll(&message.StopListen{
			Function: function,
		})
	}
}

func (i *InputCore) onStartConsume(d eventual2go.Data) {
	stream := d.(string)
	i.log.Infof("Starting to consume stream '%s'", stream)
	i.consumedStreams[stream] = nil
	i.sendToAll(&message.StartConsume{
		Stream: stream,
	})
}

func (i *InputCore) onStopConsume(d eventual2go.Data) {
	stream := d.(string)

	if _, ok := i.consumedStreams[stream]; ok {
		i.log.Infof("Stopping to consume stream '%s'", stream)
		delete(i.consumedStreams, stream)
		i.sendToAll(&message.StopConsume{
			Stream: stream,
		})
	}
}

func (i *InputCore) onAddStream(d eventual2go.Data) {
	m := d.(*message.AddStream)
	i.log.Debugf("Got value on stream '%s'", m.Name)
	if s, ok := i.streams[m.Name]; ok {
		s.Add(m.Value)
	}
}

func (i *InputCore) onStartObserve(d eventual2go.Data) {
	property := d.(string)
	i.log.Infof("Starting to observe property '%s'", property)
	i.observedProperties[property] = nil
	i.sendToAll(&message.StartObserve{
		Property: property,
	})
}

func (i *InputCore) onStopObserve(d eventual2go.Data) {
	property := d.(string)

	if _, ok := i.observedProperties[property]; ok {
		i.log.Infof("Stopping to listen to property '%s'", property)
		delete(i.observedProperties, property)
		i.sendToAll(&message.StopObserve{
			Property: property,
		})
	}
}

func (i *InputCore) onSetProperty(d eventual2go.Data) {
	m := d.(*message.SetProperty)
	i.log.Debugf("Got update for property '%s'", m.Name)
	if o, ok := i.properties[m.Name]; ok {
		o.Change(m.Value)
	}
}

func (i *InputCore) GetProperty(property string) (o *eventual2go.Observable, err error) {
	o, ok := i.properties[property]
	if !ok {
		err = fmt.Errorf("Can't get unknown property '%s'", property)
		return
	}
	return
}

func (i *InputCore) UpdateProperty(property string) (f *eventual2go.Future, err error) {
	o, err := i.GetProperty(property)
	f = o.NextChange()
	i.log.Debugf("Sending request to update property '%s'", property)
	i.mustSend(&message.GetProperty{property}, f)
	return
}

func isSetProperty(d message.Message) (is bool) {
	_, is = d.(*message.SetProperty)
	return
}

func isProperty(property string) message.MessageFilter {
	return func(d message.Message) (is bool) {
		is = d.(*message.SetProperty).Name == property
		return
	}
}

func (i *InputCore) Request(function string, callType message.CallType,
	params []byte) (result *message.ResultFuture, resultStream *message.ResultStream,
	request_uuid uuid.UUID, err error) {

	if !i.descriptor.HasFunction(function) {
		err = fmt.Errorf("Function '%s' is not in descriptor", function)
		return
	}

	req := message.NewRequest(i.UUID(), function, callType, params)

	request_uuid = req.UUID

	i.log.Debugf("New request %s %s %s", function, request_uuid, callType)
	switch callType {
	case message.CALL:
		result = i.results.FirstWhere(req.IsReply)
		i.mustSend(req, result.Future)

	case message.CALLALL:
		hasUUID := func(uuid uuid.UUID) message.ResultFilter {
			return func(r *message.Result) bool {
				return r.Request.UUID == uuid
			}
		}
		resultStream = i.results.Where(hasUUID(request_uuid))
		i.sendToAll(req)

	case message.TRIGGER:
		i.sendToOne(req)

	case message.TRIGGERALL:
		i.sendToAll(req)
	}

	return

}

func (i InputCore) ListenStream() *message.ResultStream {
	return i.results.Where(isListenResult)
}

func (i *InputCore) StartListen(function string) (err error) {

	if !i.descriptor.HasFunction(function) {
		err = fmt.Errorf("Function '%s' is not in descriptor", function)
		return
	}

	i.r.Fire(startListenEvent{}, function)

	return
}

func (i *InputCore) StopListen(function string) (err error) {

	if !i.descriptor.HasFunction(function) {
		err = fmt.Errorf("Function '%s' is not in descriptor", function)
		return
	}

	i.r.Fire(stopListenEvent{}, function)

	return
}

func (i *InputCore) GetStream(stream string) (s *eventual2go.Stream, err error) {

	if !i.descriptor.HasStream(stream) {
		err = fmt.Errorf("Stream '%s' is not in descriptor", stream)
		return
	}

	s = i.streams[stream].Stream()

	return
}

func (i *InputCore) StartConsume(stream string) (err error) {

	if !i.descriptor.HasStream(stream) {
		err = fmt.Errorf("Stream '%s' is not in descriptor", stream)
		return
	}

	i.r.Fire(startConsumeEvent{}, stream)

	return
}

func (i *InputCore) StopConsume(stream string) (err error) {

	if !i.descriptor.HasStream(stream) {
		err = fmt.Errorf("Stream '%s' is not in descriptor", stream)
		return
	}

	i.r.Fire(stopConsumeEvent{}, stream)

	return
}

func (i *InputCore) StartObservation(property string) (err error) {

	if !i.descriptor.HasProperty(property) {
		err = fmt.Errorf("Property '%s' is not in descriptor", property)
		return
	}

	i.r.Fire(startObserveEvent{}, property)

	return
}

func (i *InputCore) StopObservation(property string) (err error) {

	if !i.descriptor.HasProperty(property) {
		err = fmt.Errorf("Property '%s' is not in descriptor", property)
		return
	}

	i.r.Fire(stopObserveEvent{}, property)

	return
}
