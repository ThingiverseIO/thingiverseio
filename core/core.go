package core

import (
	"fmt"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/logging"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/eventual2go/typed_events"
	gologging "github.com/op/go-logging"
)

type Core struct {
	*eventual2go.Reactor
	config      *config.Config
	connected   *typed_events.BoolCompleter
	connections map[uuid.UUID]network.Connection
	log         *gologging.Logger
	provider    network.Providers
	tracker     network.Tracker
	shutdown    *eventual2go.Shutdown
}

func Initialize(cfg *config.Config, tracker network.Tracker, providers ...network.Provider) (c *Core, err error) {
	logging.SetupLogger(cfg)

	shutdown := eventual2go.NewShutdown()

	provider, err := network.NewProviders(cfg, providers)
	if err != nil {
		return
	}
	provider.RegisterShutdown(shutdown)

	if err = tracker.Init(cfg, provider.EncodedDetails); err != nil {
		return
	}
	shutdown.Register(tracker)

	logPrefix := fmt.Sprintf("CORE %s", cfg.Internal.UUID)

	c = &Core{
		Reactor:     eventual2go.NewReactor(),
		config:      cfg,
		connected:   typed_events.NewBoolCompleter(),
		connections: map[uuid.UUID]network.Connection{},
		log:         logging.CreateLogger(logPrefix, cfg),
		provider:    provider,
		tracker:     tracker,
		shutdown:    shutdown,
	}

	c.Reactor.React(connectEvent{}, c.onConnection)

	c.Reactor.AddStream(leaveEvent{}, tracker.Leaving().Stream)
	c.Reactor.React(leaveEvent{}, c.onLeave)

	c.Reactor.AddStream(endEvent{}, c.provider.Messages().Where(network.OfType(message.END)).Stream)
	c.Reactor.React(endEvent{}, c.onEnd)

	c.Reactor.React(shutdownEvent{}, c.onShutdown)

	c.log.Info("Started")
	return
}

func (c Core) Connected() (is bool) {
	return c.ConnectedFuture().Completed()
}

func (c Core) ConnectedFuture() (is *typed_events.BoolFuture) {
	return c.connected.Future()
}

func (c *Core) onConnection(d eventual2go.Data) {
	conn := d.(network.Connection)
	c.connections[conn.UUID] = conn
	c.log.Infof("Connected to %s", conn.UUID)
	c.log.Info("BLUB", c.connected.Completed())
	if !c.connected.Completed() {
	c.log.Debug("LA1")
		c.connected.Complete(true)
	c.log.Debug("LA2")
		c.tracker.StopAdvertisment()
		c.log.Info("Connected")
	}
	c.Reactor.Fire(afterConnectedEvent{}, conn.UUID)
	c.log.Debug("BLA")
}

func (c *Core) onEnd(d eventual2go.Data) {
	m := d.(network.Message)
	c.log.Info("Received and from", m.Sender)
	c.removePeer(m.Sender)
}

func (c *Core) onLeave(d eventual2go.Data) {
	uuid := d.(uuid.UUID)
	c.log.Info("Peer left", uuid)
	c.removePeer(uuid)
}

func (c Core) onShutdown(d eventual2go.Data) {
	c.log.Info("Shutting down")
	m := &message.End{}
	for _, conn := range c.connections {
		conn.Send(m)
		conn.Close()
	}
	c.shutdown.Do(nil)
	d.(*eventual2go.Completer).Complete(nil)
	c.log.Info("Shutdown complete")
}

func (c *Core) removePeer(uuid uuid.UUID) {
	delete(c.connections, uuid)
	if len(c.connections) == 0 {
		c.connected = typed_events.NewBoolCompleter()
		c.tracker.StartAdvertisment()
		c.log.Info("Disconnected")
	}
	c.Reactor.Fire(afterPeerRemovedEvent{}, uuid)
}

func (c Core) Run() {
	c.tracker.StartAdvertisment()
}

func (c Core) SendToAll(m message.Message) {
	for _, conn := range c.connections {
		conn.Send(m)
	}
}

func (c Core) Shutdown() {
	cmp := eventual2go.NewCompleter()
	c.Reactor.Fire(shutdownEvent{}, cmp)
	cmp.Future().WaitUntilComplete()
}

func (c Core) UUID() uuid.UUID {
	return c.config.Internal.UUID
}
