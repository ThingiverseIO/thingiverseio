package thingiverseio

import (
	"fmt"
	"log"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/service/connection"
	"github.com/ThingiverseIO/thingiverseio/service/manager"
	"github.com/ThingiverseIO/thingiverseio/service/messages"
	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/eventual2go/typed_events"
)

// Output is a ThingiverseIO node which exports functionality to the ThingiverseIO network.
type Output struct {
	cfg       *config.Config
	m         *manager.Manager
	listeners map[string]map[config.UUID]interface{}
	logger    *log.Logger
	r         *eventual2go.Reactor
	requests  *messages.RequestStream
}

// NewOutput creates a new Output instance for a given service descriptor. Configuration is automatically determined by the thingiversio/config package.
func NewOutput(desc string) (o *Output, err error) {
	var d Descriptor
	d, err = ParseDescriptor(desc)
	if err == nil {
		o, err = NewOutputFromConfig(config.Configure(true, d.AsTagSet()))
	}
	return
}

// NewOutputFromConfig creates a new Output instance for a given configuration.
func NewOutputFromConfig(cfg *config.Config) (o *Output, err error) {
	m, err := manager.New(cfg)
	o = &Output{
		cfg:       cfg,
		m:         m,
		requests:  &messages.RequestStream{m.MessagesOfType(messages.REQUEST).Transform(connection.ToMessage)},
		listeners: map[string]map[config.UUID]interface{}{},
		logger:    log.New(cfg.Logger(), fmt.Sprintf("%s OUTPUT ", cfg.UUID()), 0),
		r:         eventual2go.NewReactor(),
	}

	o.logger.Println("Launching with tagset", cfg.Tags())

	o.r.React(startListenEvent{}, o.onListen)
	o.r.AddStream(startListenEvent{}, m.MessagesOfType(messages.LISTEN).Stream)

	o.r.React(stopListenEvent{}, o.onStopListen)
	o.r.AddStream(stopListenEvent{}, m.MessagesOfType(messages.STOPLISTEN).Stream)

	o.r.React(leaveEvent{}, o.onPeerGone)

	o.r.React(replyEvent{}, o.deliverResult)
	return
}

// UUID returns the UUID of a Output instance.
func (o *Output) UUID() config.UUID {
	return o.cfg.UUID()
}

// Interface returns the address of the interface the Output is using.
func (o *Output) Interface() string {
	return o.cfg.Interfaces()[0]
}

// Remove shuts down the Output.
func (o *Output) Remove() (errs []error) {
	errs = o.m.Shutdown()
	o.r.Shutdown(nil)
	return
}

// Run starts the Output creating all connections and starting service discovery.
func (o *Output) Run() {
	o.m.Run()
}

// Connected returns a eventual2go.Future which gets completed when a suitable Input is discovered.
func (o *Output) Connected() *typed_events.BoolFuture {
	return o.m.Connected().FirstWhere(func(b bool) bool { return b })
}

// Disconnected returns a eventual2go.Future which gets completed when the last suitable Input is removed from the network.
func (o *Output) Disconnected() *typed_events.BoolFuture {
	return o.m.Connected().FirstWhereNot(func(b bool) bool { return b })
}

// Reply reponds the given output parameter to all interested Inputs of a given request.
func (o *Output) Reply(r *messages.Request, params interface{}) {
	res := messages.NewResult(o.cfg.UUID(), r, params)
	o.r.Fire(replyEvent{}, res)
}

// ReplyEncoded does the same as Reply, but takes already encoded return parameters. Used mainly by the shared library.
func (o *Output) ReplyEncoded(r *messages.Request, params []byte) {
	res := messages.NewEncodedResult(o.cfg.UUID(), r, params)
	o.r.Fire(replyEvent{}, res)
}

// Emit acts like a ThingiverseIO Trigger, which is initiated by the Output.
func (o *Output) Emit(function string, inparams interface{}, outparams interface{}) {
	req := messages.NewRequest(o.cfg.UUID(), function, messages.TRIGGER, inparams)
	o.Reply(req, outparams)
}

// EmitEncoded does the same as Emit, but takes already encoded return parameters. Used mainly by the shared library.
func (o *Output) EmitEncoded(function string, inparams []byte, outparams []byte) {
	req := messages.NewEncodedRequest("", function, messages.TRIGGER, inparams)
	o.ReplyEncoded(req, outparams)
}

// Requests returns a RequestStream, which delivers incoming requests. Although multiple listeners can be registered, multiple replies to one request can lead to undefined behaviour.
func (o *Output) Requests() *messages.RequestStream {
	return o.requests
}

func (o *Output) onListen(d eventual2go.Data) {
	m := d.(connection.Message)
	l := messages.Unflatten(m.Payload)
	f := l.(*messages.Listen).Function
	o.logger.Println("New Listener", m.Sender, f)
	_, ok := o.listeners[f]
	if ok {
		o.listeners[f][m.Sender] = nil
	} else {
		o.listeners[f] = map[config.UUID]interface{}{m.Sender: nil}
	}
	o.r.AddFuture(leaveEvent{}, o.m.PeerLeave(m.Sender).Future)
}

func (o *Output) onStopListen(d eventual2go.Data) {
	m := d.(connection.Message)
	l := messages.Unflatten(m.Payload)
	f := l.(*messages.StopListen).Function
	o.removePeerListen(m.Sender, f)
}

func (o *Output) onPeerGone(d eventual2go.Data) {
	uuid := d.(config.UUID)
	for f := range o.listeners {
		o.removePeerListen(uuid, f)
	}
}

func (o *Output) removePeerListen(uuid config.UUID, f string) {
	_, ok := o.listeners[f]
	if !ok {
		return
	}
	_, ok = o.listeners[f][uuid]
	if ok {
		delete(o.listeners[f], uuid)
	}
}

func (o *Output) deliverResult(d eventual2go.Data) {
	result := d.(*messages.Result)
	o.logger.Println("Delivering result", result.Request.Function, result.Request.CallType)

	switch result.Request.CallType {
	case messages.CALL, messages.CALLALL:
		o.logger.Println("Delivering to", result.Request.Input)
		o.m.SendTo(result.Request.Input, result)

	case messages.TRIGGER, messages.TRIGGERALL:
		if ls, ok := o.listeners[result.Request.Function]; ok {
			for uuid := range ls {
				o.m.SendTo(uuid, result)
			}
		}
	}
}
