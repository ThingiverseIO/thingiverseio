package thingiverseio

import (
	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/core"
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/joernweissenborn/eventual2go/typed_events"
)

// Input is a ThingiverseIO node which imports functionality from the ThingiverseIO network.
type Input struct {
	core core.InputCore
}

// NewInput creates a new Input instance for a given service descriptor. Configuration is automatically determined by the thingiversio/config package.
func NewInput(desc string) (i *Input, err error) {
	i, err = NewInputFromConfig(desc, config.Configure())
	return
}

// NewInputFromConfig creates a new Input instance for a given configuration.
func NewInputFromConfig(desc string, cfg *config.UserConfig) (i *Input, err error) {
	var d descriptor.Descriptor
	if d, err = descriptor.Parse(desc); err != nil {
		return
	}

	tracker, provider := getDefaultBackends()

	core, err := core.NewInputCore(d, cfg, tracker, provider...)
	i = &Input{
		core: core,
	}

	return
}

// Remove shuts down the Input.
func (i *Input) Remove() {
	i.core.Shutdown()
}

// Run starts service discovery.
func (i *Input) Run() {
	i.core.Run()
}

// UUID returns the UUID of an Input instance.
func (i *Input) UUID() uuid.UUID {
	return i.core.UUID()
}

// Connected returns true if the Input instance is connected to at least 1 suitable Output, otherwise false.
func (i *Input) Connected() bool {
	return i.core.Connected()
}

// ConnectedFuture returns a eventual2go.Future which gets completed when the first suitable Output is connected.
func (i *Input) ConnectedFuture() *typed_events.BoolFuture {
	return i.core.ConnectedFuture()
}

// Call executes a ThingiverseIO Call and returns a ResultFuture, which gets completed if a suitable output reponds.
func (i *Input) Call(function string, parameter interface{}) (result *message.ResultFuture, err error) {
	data, err := encode(parameter)
	if err != nil {
		return
	}
	result, _, _, err = i.core.Request(function, message.CALL, data)
	return
}

// CallAll executes a ThingiverseIO CallAll and returns the Requests UUID. A ResultStreamController must be provided by the user, who must also handle it's lifetime. The current implementation is incomplete. If you close the StreamController and a response is received after, the library might crash. Use with care.
func (i *Input) CallAll(function string, parameter interface{}) (results *message.ResultStream, err error) {
	data, err := encode(parameter)
	if err != nil {
		return
	}
	_, results, _, err = i.core.Request(function, message.CALLALL, data)
	return
}

// Trigger executes a ThingiverseIO Trigger.
func (i *Input) Trigger(function string, parameter interface{}) (err error) {
	data, err := encode(parameter)
	if err != nil {
		return
	}
	_, _, _, err = i.core.Request(function, message.TRIGGER, data)
	return
}

// TriggerAll executes a ThingiverseIO TriggerAll.
func (i *Input) TriggerAll(function string, parameter interface{}) (err error) {
	data, err := encode(parameter)
	if err != nil {
		return
	}
	_, _, _, err = i.core.Request(function, message.TRIGGERALL, data)
	return
}

// StartListen starts listening on the given function.
func (i *Input) StartListen(function string) {
	i.core.StartListen(function)
}

// StopListen stops listening on the given function.
func (i *Input) StopListen(function string) {
	i.core.StopListen(function)
}

// ListenResults returns a ResultStream to receive results of Trigger or TriggerAll function calls.
func (i *Input) ListenResults() *message.ResultStream {
	return i.core.ListenStream()
}
