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

type pendingRequest struct {
	*message.ResultCompleter
	output  uuid.UUID
	request *message.Request
}

type InputCore struct {
	*Core
	descriptor      descriptor.Descriptor
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
		pendingRequests: map[uuid.UUID]*pendingRequest{},
	}

	i.results = &message.ResultStream{
		Stream: i.provider.Messages().
			Where(network.OfType(message.RESULT)).
			Transform(network.ToMessage),
	}

	i.tracker.Arrivals().Where(i.isInterestingArrival).Listen(i.onArrival)

	i.Reactor.React(requestEvent{}, i.onRequest)

	return

}

func (i *InputCore) deliverRequest(req *message.Request) {

	switch req.CallType {

	case message.CALL, message.TRIGGER:
		for id, conn := range i.connections {
			if req.CallType == message.CALL {
				i.pendingRequests[req.UUID].output = id
			}
			conn.Send(req)
			return
		}

	case message.TRIGGERALL, message.CALLALL:

	}
}

func (i *InputCore) Request(function string, callType message.CallType,
	params []byte) (result *message.ResultFuture, uuid uuid.UUID, err error) {

	if !i.descriptor.HasFunction(function) {
		err = fmt.Errorf("Function '%s' is not in descriptor", function)
		return
	}

	req := message.NewRequest(i.UUID(), function, callType, params)

	uuid = req.UUID

	if callType == message.CALL {
		preq := &pendingRequest{
			ResultCompleter: message.NewResultCompleter(),
			request:         req,
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

	i.Reactor.Fire(requestEvent{}, req)

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

	return
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

	i.log.Debugf("Trying to delivering result %s", req.UUID)

	if i.Connected() {
		i.deliverRequest(req)
	}
}

func (i *InputCore) onReply(d eventual2go.Data) {
	res := d.(*message.Result)
	delete(i.pendingRequests, res.Request.UUID)
}

func (i InputCore) Pending() map[uuid.UUID]*pendingRequest {
	return i.pendingRequests
}
