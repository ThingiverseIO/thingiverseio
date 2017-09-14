package core

import (
	"fmt"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ThingiverseIO/logger"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/joernweissenborn/eventual2go"
)

type core struct {
	*eventual2go.Reactor
	config           *config.Config
	connected        *eventual2go.Completer
	descriptor       descriptor.Descriptor
	disconnected     *eventual2go.Completer
	connections      map[uuid.UUID]network.Connection
	log              *logger.Logger
	mustSendRegister map[message.Message]uuid.UUID
	provider         network.Providers
	tracker          network.Tracker
	shutdown         *eventual2go.Shutdown
	properties       properties
}

func initCore(desc descriptor.Descriptor, cfg *config.Config, tracker network.Tracker, providers ...network.Provider) (c *core, err error) {

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
	writer := cfg.User.GetLogger()
	if err !=nil{
		return
	}
	logger.SetDefaultBackend(logger.NewMultiWriteBackend(writer))
	logPrefix := fmt.Sprintf("TVIO %s", cfg.Internal.UUID)

	c = &core{
		Reactor:          eventual2go.NewReactor(),
		config:           cfg,
		connected:        eventual2go.NewCompleter(),
		descriptor:       desc,
		disconnected:     eventual2go.NewCompleter(),
		connections:      map[uuid.UUID]network.Connection{},
		log:              logger.New(logPrefix),
		mustSendRegister: map[message.Message]uuid.UUID{},
		provider:         provider,
		tracker:          tracker,
		shutdown:         shutdown,
		properties:       newProperties(desc),
	}

	c.Reactor.React(connectEvent{}, c.onConnection)
	c.Reactor.React(connectEvent{}, c.onConnection)

	c.Reactor.AddStream(leaveEvent{}, tracker.Leaving().Stream)
	c.Reactor.React(leaveEvent{}, c.onLeave)
	c.Reactor.React(mustSendEvent{}, c.onMustSend)

	c.Reactor.AddStream(endEvent{}, c.provider.Messages().Where(network.OfType(message.END)).Stream)
	c.Reactor.React(endEvent{}, c.onEnd)

	c.Reactor.OnShutdown(c.onShutdown)

	c.log.Info("Started")
	return
}

func (c *core) Connected() (is bool) {
	c.Reactor.Lock()
	defer c.Reactor.Unlock()
	return c.connected.Future().Completed()
}

func (c *core) ConnectedFuture() (is *eventual2go.Future) {
	c.Lock()
	defer c.Unlock()
	return c.connected.Future()
}

func (c *core) DisconnectedFuture() (is *eventual2go.Future) {
	c.Lock()
	defer c.Unlock()
	return c.disconnected.Future()
}

func (c *core) Interface() string {
	return c.config.User.Interface
}

func (c *core) Properties() (properties []string) {
	for p, _ := range c.properties {
		properties = append(properties, p)
	}
	return
}

func (c *core) onConnection(d eventual2go.Data) {
	conn := d.(network.Connection)
	c.connections[conn.UUID] = conn
	c.log.Infof("Connected to %s", conn.UUID)
	if !c.connected.Completed() {
		c.connected.Complete(true)
		c.tracker.StopAdvertisment()
		c.log.Info("Connected")
		for m, id := range c.mustSendRegister {
			if id.IsEmpty() {
				c.mustSendRegister[m] = conn.UUID
				conn.Send(m)
			}
		}
	}
	c.Reactor.Fire(afterConnectedEvent{}, conn.UUID)
}

func (c *core) onEnd(d eventual2go.Data) {
	m := d.(network.Message)
	c.log.Info("Received END from", m.Sender)
	c.removePeer(m.Sender)
}

func (c *core) onLeave(d eventual2go.Data) {
	uuid := d.(uuid.UUID)
	c.log.Info("Peer left", uuid)
	c.removePeer(uuid)
}

func (c *core) onShutdown(d eventual2go.Data) {
	c.log.Info("Shutting down")
	m := &message.End{}
	for _, conn := range c.connections {
		conn.Send(m)
		conn.Close()
	}
	c.shutdown.Do(nil)
	d.(*eventual2go.Completer).Complete(nil)
}

func (c *core) removePeer(uuid uuid.UUID) {

	if conn, ok := c.connections[uuid]; ok {
		c.log.Debug("Closing connections to peer", uuid)
		conn.Close()
		delete(c.connections, uuid)
		if len(c.connections) == 0 {
			c.connected = eventual2go.NewCompleter()
			c.disconnected.Complete(true)
			c.disconnected = eventual2go.NewCompleter()
			c.tracker.StartAdvertisment()
			c.log.Info("Disconnected")
		}
		for m, id := range c.mustSendRegister {
			if id == uuid {
				c.log.Debug("Removed peer had pending message")
				c.Fire(mustSendEvent{}, m)
			}
		}
		c.Fire(afterPeerRemovedEvent{}, uuid)
	}
}

func (c *core) Run() {
	c.tracker.StartAdvertisment()
}

func (c *core) sendToOne(m message.Message) {
	c.log.Debug("Trying to send message to one peer:", m.GetType())
	for id, conn := range c.connections {
		c.log.Debug("Sending message to ", id)
		conn.Send(m)
		return
	}
	c.log.Debug("Cant send, no connections")
}

func (c *core) sendToAll(m message.Message) {
	c.log.Debug("Trying to send message to all peers:", m.GetType())
	for id, conn := range c.connections {
		c.log.Debug("Sending message to ", id)
		conn.Send(m)
	}
}

func (c *core) Shutdown() {
	cmp := eventual2go.NewCompleter()
	c.Reactor.Shutdown(cmp)
	cmp.Future().WaitUntilComplete()
	c.log.Info("Shutdown complete")
}

func (c *core) UUID() uuid.UUID {
	return c.config.Internal.UUID
}

func (c *core) mustSend(m message.Message, recv *eventual2go.Future) {
	recv.Then(c.onRecv(m))
	c.Fire(mustSendEvent{}, m)
}

func (c *core) onMustSend(d eventual2go.Data) {
	m := d.(message.Message)
	c.mustSendRegister[m] = uuid.Empty()
	c.log.Debug("Trying to send must message", m.GetType())
	for id, conn := range c.connections {
		c.mustSendRegister[m] = id
		c.log.Debug("Sending must message to", id)
		conn.Send(m)
		return
	}
}

func (c *core) onRecv(m message.Message) eventual2go.CompletionHandler {
	return func(eventual2go.Data) eventual2go.Data {
		c.Lock()
		defer c.Unlock()
		delete(c.mustSendRegister, m)
		return nil
	}
}
