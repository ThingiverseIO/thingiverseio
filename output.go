package thingiverseio

import (
	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/core"
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ThingiverseIO/uuid"
	"github.com/joernweissenborn/eventual2go/typedevents"
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

	tracker, provider := core.DefaultBackends()

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
func (o *Output) Connected() bool {
	return o.core.Connected()
}

// ConnectedObservable returns a eventual2go/typedevents.BoolObservable which represents the connection state.
func (o *Output) ConnectedObservable() *typedevents.BoolObservable {
	return o.core.ConnectedObservable()
}

// WaitUntilConnected waits until the output is connected.
func (o *Output) WaitUntilConnected() {
	f := o.ConnectedObservable().Stream().First()
	if o.Connected() {
		return
	}
	f.WaitUntilComplete()
}

// Reply reponds the given output parameter to all interested Inputs of a given request.
func (o *Output) Reply(request *Request, parameter interface{}) (err error) {
	data, err := encode(parameter)
	if err != nil {
		return
	}
	o.core.Reply(request, data)
	return
}

// Emit acts like a ThingiverseIO Trigger, which is initiated by the Output.
func (o *Output) Emit(function string, inparams interface{}, outparams interface{}) (err error) {
	inp, err := encode(inparams)
	if err != nil {
		return
	}
	outp, err := encode(outparams)
	if err != nil {
		return
	}
	err = o.core.Emit(function, inp, outp)
	return
}

// Requests returns a RequestStream, which delivers incoming requests. Although multiple listeners can be registered, multiple replies to one request can lead to undefined behaviour.
func (o *Output) Requests() *RequestStream {
	return &RequestStream{o.core.RequestStream().Stream}
}

// RequestsWhereFunction returns a RequestStream only for the given function.
func (o *Output) RequestsWhereFunction(function string) *RequestStream {
	return (&RequestStream{o.core.RequestStream().Stream}).Where(filterRequests(function))
}

func filterRequests(function string) RequestFilter {
	return func(r *Request) bool {return r.Function==function}
}

// SetProperty sets the value of a property.
func (o *Output) SetProperty(property string, value interface{}) (err error) {
	v, err := encode(value)
	if err != nil {
		return
	}
	err = o.core.SetProperty(property, v)
	return
}

// AddStream adds a value on a stream.
func (o *Output) AddStream(stream string, value interface{}) (err error) {
	v, err := encode(value)
	if err != nil {
		return
	}
	err = o.core.AddStream(stream, v)
	return
}
