package thingiverseio

import (
	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/core"
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/joernweissenborn/eventual2go/typed_events"
)

// Output is a ThingiverseIO node which exports functionality to the ThingiverseIO network.
type Output struct {
	core core.OutputCore
}

// NewOutput creates a new Output instance for a given service descriptor. Configuration is automatically determined by the thingiversio/config package.
func NewOutput(desc string) (o *Output, err error) {
	o, err = NewOutputFromConfig(desc, config.Configure())
	return
}

// NewOutputFromConfig creates a new Output instance for a given configuration.
func NewOutputFromConfig(desc string, cfg *config.UserConfig) (o *Output, err error) {
	var d descriptor.Descriptor
	if d, err = descriptor.Parse(desc); err != nil {
		return
	}

	tracker, provider := getDefaultBackends()

	core, err := core.NewOutputCore(d, cfg, tracker, provider...)
	o = &Output{
		core: core,
	}
	return
}

// UUID returns the UUID of a Output instance.
func (o *Output) UUID() uuid.UUID {
	return o.core.UUID()
}

// Remove shuts down the Output.
func (o *Output) Remove() {
	o.core.Shutdown()
}

// Run starts the Output creating all connections and starting service discovery.
func (o *Output) Run() {
	o.core.Run()
}

// Connected returns true if at least 1 Input is connected.
func (o *Output) Connected() *typed_events.BoolFuture {
	return o.core.ConnectedFuture()
}

// ConnectedFuture returns a eventual2go.Future which gets completed when a suitable Input is connected.
func (o *Output) ConnectedFuture() *typed_events.BoolFuture {
	return o.core.ConnectedFuture()
}

// Reply reponds the given output parameter to all interested Inputs of a given request.
func (o *Output) Reply(request *message.Request, parameter interface{}) (err error) {
	data, err := encode(parameter)
	if err != nil {
		return
	}
	o.core.Reply(request, data)
	return
}

// Emit acts like a ThingiverseIO Trigger, which is initiated by the Output.
func (o *Output) Emit(function string, inparams interface{}, outparams interface{}) (err error) {
	data, err := encode(inparams)
	if err != nil {
		return
	}
	req := message.NewRequest(o.UUID(), function, message.TRIGGER, data)
	o.Reply(req, outparams)
	return
}

// Requests returns a RequestStream, which delivers incoming requests. Although multiple listeners can be registered, multiple replies to one request can lead to undefined behaviour.
func (o *Output) Requests() *message.RequestStream {
	return o.core.RequestStream()
}
