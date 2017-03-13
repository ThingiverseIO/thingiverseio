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

func isListenResult(r *message.Result) bool {
	return r.Request.CallType == message.TRIGGER || r.Request.CallType == message.TRIGGERALL
}

type pendingRequest struct {
	*message.ResultCompleter
	Output  uuid.UUID
	Request *message.Request
}

type InputCore struct {
	*Core
	descriptor      descriptor.Descriptor
	listenFunctions map[string]interface{}
	pendingRequests map[uuid.UUID]*pendingRequest
	results         *message.ResultStream
}

func NewInputCore(desc descriptor.Descriptor, usrCfg *config.UserConfig,
	tracker network.Tracker, provider ...network.Provider) (i InputCore, err error) {

	tags := desc.AsTagset()
	tags.Merge(usrCfg.Tags)

	intCfg := config.NewInternalConfig(false, tags)

	c, err := Initialize(&config.Config{
		Internal: intCfg,
		User:     usrCfg,
	}, tracker, provider...)

	i = InputCore{
		Core:            c,
		descriptor:      desc,
		listenFunctions: map[string]interface{}{},
		pendingRequests: map[uuid.UUID]*pendingRequest{},
	}

	i.results = &message.ResultStream{
		Stream: i.provider.Messages().
			Where(network.OfType(message.RESULT)).
			Transform(network.ToMessage),
	}

	i.tracker.Arrivals().Where(i.isInterestingArrival).Listen(i.onArrival)

	i.Reactor.React(afterConnectedEvent{}, i.onAfterConnected)

	i.Reactor.React(afterPeerRemovedEvent{}, i.onAfterPeerRemoved)

	i.Reactor.React(replyEvent{}, i.onReply)

	i.Reactor.React(requestEvent{}, i.onRequest)

	i.Reactor.React(startListenEvent{}, i.onStartListen)

	i.Reactor.React(stopListenEvent{}, i.onStopListen)

	return

}

func (i *InputCore) deliverRequest(req *message.Request) {

	for id, conn := range i.connections {
		i.log.Debugf("Delivering request to %s", id)
		conn.Send(req)
		switch req.CallType {
		case message.CALL, message.TRIGGER:
			if req.CallType == message.CALL {
				i.pendingRequests[req.UUID].Output = id
			}
			return

		case message.TRIGGERALL, message.CALLALL:
		}

	}
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

	return
}

func (i InputCore) ListenStream() *message.ResultStream {
	return i.results.Where(isListenResult)
}

func (i *InputCore) onAfterConnected(d eventual2go.Data) {
	uuid := d.(uuid.UUID)

	for function := range i.listenFunctions {
		i.log.Debugf("Informing %s about function listeng to '%s'", uuid, function)
		i.connections[uuid].Send(&message.StartListen{
			Function: function,
		})
	}

	for id, pending := range i.Pending() {
		if pending.Output.IsEmpty() {
			i.log.Debugf("Trying to delivering request %s", id)
			i.deliverRequest(pending.Request)
		}
	}
}

func (i *InputCore) onAfterPeerRemoved(d eventual2go.Data) {
	id := d.(uuid.UUID)
	for reqeuestId, pending := range i.Pending() {
		if pending.Output == id {
			i.log.Debug("Removed peer had pending request", reqeuestId)
			pending.Output = uuid.Empty()
			if i.connected.Completed() {
				i.deliverRequest(pending.Request)
			}
		}
	}
}

func (i InputCore) onArrival(a network.Arrival) {
	i.log.Debug("Peer arrived")
	conn, err := i.provider.Connect(a.Details, a.UUID)
	if err != nil {
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

	if next.WaitUntilTimeout(1 * time.Second) {
		i.log.Debug("Received HELLOOK")
		if next.Result().Decode().(*message.HelloOk).Have {
			i.log.Debug("First Tag is supported, checking all tags")
			for _, tag := range i.config.Internal.Tags.AsArray() {

				i.log.Debug("Checking for tag", tag)
				next = in.FirstWhere(network.OfType(message.HAVE))

				// Send DoHave
				conn.Send(&message.DoHave{
					Tag: tag,
				})

				if next.WaitUntilTimeout(1 * time.Second) {
					i.log.Debug("Got message HAVE")
					if !next.Result().Decode().(*message.Have).Have {
						// Send End

						i.log.Debug("Peer not supported, aborting")
						conn.Send(&message.End{})
						conn.Close()
						return
					}
				} else {
					conn.Close()
					return
				}
			}

			// Send Connect

			conn.Send(&message.Connect{})
			i.Fire(connectEvent{}, conn)
			return
		}
		i.log.Debug("Peer not supported, aborting")

	} else {
		i.log.Debug("Timeout")
	}

	conn.Close()

}

func (i *InputCore) onRequest(d eventual2go.Data) {
	req := d.(*message.Request)

	i.log.Debugf("Trying to delivering request %s", req.UUID)

	if i.connected.Completed() {
		i.deliverRequest(req)
	} else {
		i.log.Debug("Can't deliver, no connections.")
	}
}

func (i *InputCore) onReply(d eventual2go.Data) {
	res := d.(*message.Result)
	i.log.Debug("Got reply for", res.Request.UUID)
	delete(i.pendingRequests, res.Request.UUID)
}

func (i *InputCore) onStartListen(d eventual2go.Data) {
	function := d.(string)
	i.log.Infof("Starting to listen to function '%s'", function)
	i.listenFunctions[function] = nil
	i.SendToAll(&message.StartListen{
		Function: function,
	})
}

func (i *InputCore) onStopListen(d eventual2go.Data) {
	function := d.(string)
	i.log.Infof("Stopping to listen to function '%s'", function)

	if _, ok := i.listenFunctions[function]; ok {
		delete(i.listenFunctions, function)
		i.SendToAll(&message.StopListen{
			Function: function,
		})
	}
}

func (i InputCore) Pending() map[uuid.UUID]*pendingRequest {
	return i.pendingRequests
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

	if callType == message.CALL {
		preq := &pendingRequest{
			ResultCompleter: message.NewResultCompleter(),
			Request:         req,
		}

		result = preq.Future()
		reply := i.results.FirstWhere(req.IsReply).Future
		preq.CompleteOnFuture(reply)

		// TODO: Avoid Locking the reactor here
		i.Reactor.Lock()
		i.pendingRequests[req.UUID] = preq
		i.Reactor.Unlock()
		i.Reactor.AddFuture(replyEvent{}, reply)
	}

	if callType == message.CALLALL {

		hasUUID := func(uuid uuid.UUID) message.ResultFilter {
			return func(r *message.Result) bool {
				return r.Request.UUID == uuid
			}
		}
		resultStream = i.results.Where(hasUUID(request_uuid))
	}

	i.Reactor.Fire(requestEvent{}, req)

	return

}

func (i *InputCore) StartListen(function string) (err error) {

	if !i.descriptor.HasFunction(function) {
		err = fmt.Errorf("Function '%s' is not in descriptor", function)
		return
	}

	i.Reactor.Fire(startListenEvent{}, function)

	return
}

func (i *InputCore) StopListen(function string) {

	i.Reactor.Fire(stopListenEvent{}, function)

	return
}
