package core

import (
	"fmt"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/logging"
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/joernweissenborn/eventual2go"
	gologging "github.com/op/go-logging"
)

type Core struct {
	*eventual2go.Reactor
	config      *config.Config
	connected   *eventual2go.Completer
	connections map[uuid.UUID]network.Connection
	log         *gologging.Logger
	provider    network.Providers
	tracker     network.Tracker
}

func Initialize(cfg *config.Config, tracker network.Tracker, providers ...network.Provider) (c *Core, err error) {
	logging.SetupLogger(cfg)

	provider, err := network.NewProviders(cfg, providers)
	if err != nil {
		return
	}

	if err = tracker.Init(cfg, provider.EncodedDetails); err != nil {
		return
	}
	logPrefix := fmt.Sprintf("CORE %s", cfg.Internal.UUID)

	c = &Core{
		Reactor:     eventual2go.NewReactor(),
		config:      cfg,
		connected:   eventual2go.NewCompleter(),
		connections: map[uuid.UUID]network.Connection{},
		log:         logging.CreateLogger(logPrefix, cfg),
		provider:    provider,
		tracker:     tracker,
	}

	c.Reactor.React(connectEvent{}, c.onConnection)

	c.Reactor.AddStream(leaveEvent{}, tracker.Leaving().Stream)
	c.Reactor.React(leaveEvent{}, c.onLeave)

	c.log.Info("Started")
	return
}

func (c Core) afterConnected(){}

func (c Core) Connected() (is bool) {
	return c.ConnectedFuture().Completed()
}

func (c Core) ConnectedFuture() (is *eventual2go.Future) {
	return c.connected.Future()
}

func (c *Core) onConnection(d eventual2go.Data) {
	conn := d.(network.Connection)
	c.connections[conn.UUID] = conn
	c.log.Infof("Connected to %s", conn.UUID)
	if !c.connected.Completed() {
		c.connected.Complete(nil)
		c.log.Info("Connected")
		c.afterConnected()
	}
}

func (c *Core) onLeave(d eventual2go.Data) {
	uuid := d.(uuid.UUID)
	c.log.Info("Peer left", uuid)
	delete(c.connections, uuid)
	if len(c.connections) == 0 {
		c.connected = eventual2go.NewCompleter()
		c.log.Info("Disconnected")
	}
}

func (c Core) UUID() uuid.UUID {
	return c.config.Internal.UUID
}
