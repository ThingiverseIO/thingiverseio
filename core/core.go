package core

import (
	"fmt"
	"time"

	"github.com/ThingiverseIO/logger"
	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/eventual2go/typedevents"
)

var (
	connectionTimeout = 1 * time.Second
)

type core struct {
	r *eventual2go.Reactor
	config           *config.Config
	connected        *typedevents.BoolObservable
	descriptor       descriptor.Descriptor
	connections      map[uuid.UUID]network.Connection
	log              *logger.Logger
	mustSendRegister map[message.Message]uuid.UUID
	provider         network.Providers
	tracker          network.Tracker
	shutdown         *eventual2go.Shutdown
	properties       properties
	streams          streams
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
	if err != nil {
		return
	}
	logger.SetDefaultBackend(logger.NewMultiWriteBackend(writer))
	logPrefix := fmt.Sprintf("TVIO %s", cfg.Internal.UUID)

	c = &core{
		r:          eventual2go.NewReactor(),
		config:           cfg,
		connected:        typedevents.NewBoolObservable(false),
		descriptor:       desc,
		connections:      map[uuid.UUID]network.Connection{},
		log:              logger.New(logPrefix).SetDebug(cfg.User.Debug),
		mustSendRegister: map[message.Message]uuid.UUID{},
		provider:         provider,
		tracker:          tracker,
		shutdown:         shutdown,
		properties:       newProperties(desc),
		streams:          newStreams(desc),
	}
	c.log.Init("Core staring up")
	c.log.Init("Configuration \n", c.config.User)
	c.r.React(connectEvent{}, c.onConnection)

	c.r.AddStream(leaveEvent{}, tracker.Leaving().Stream)
	c.r.React(leaveEvent{}, c.onLeave)
	c.r.React(mustSendEvent{}, c.onMustSend)

	c.r.AddStream(endEvent{}, c.provider.Messages().Where(network.OfType(message.END)).Stream)
	c.r.React(endEvent{}, c.onEnd)

	c.r.OnShutdown(c.onShutdown)
	c.log.Initf("Tagset is: %s", cfg.Internal.Tags)
	c.log.Init("Core initialized")
	return
}

func (c *core) Connected() (is bool) {
	return c.connected.Value()
}

func (c *core) ConnectedObservable() (is *typedevents.BoolObservable) {
	return c.connected
}

func (c *core) Interface() string {
	return c.config.User.Interface
}

func (c *core) Properties() (properties []string) {
	for p := range c.properties {
		properties = append(properties, p)
	}
	return
}

func (c *core) onConnection(d eventual2go.Data) {
	conn := d.(network.Connection)
	c.connections[conn.UUID] = conn
	c.log.Infof("Connected to %s", conn.UUID)
	if !c.connected.Value() {
		c.connected.Change(true)
		c.tracker.StopAdvertisment()
		c.log.Info("Connected")
		for m, id := range c.mustSendRegister {
			if id.IsEmpty() {
				c.mustSendRegister[m] = conn.UUID
				conn.Send(m)
			}
		}
	}
	c.r.Fire(afterConnectedEvent{}, conn)
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
			c.connected.Change(false)
			c.tracker.StartAdvertisment()
			c.log.Info("Disconnected")
		}
		for m, id := range c.mustSendRegister {
			if id == uuid {
				c.log.Debug("Removed peer had pending message")
				c.r.Fire(mustSendEvent{}, m)
			}
		}
		c.r.Fire(afterPeerRemovedEvent{}, uuid)
	}
}

func (c *core) Run() {
	c.log.Info("Core starting")
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
	c.r.Shutdown(cmp)
	cmp.Future().WaitUntilComplete()
	c.log.Info("Shutdown complete")
}

func (c *core) UUID() uuid.UUID {
	return c.config.Internal.UUID
}

func (c *core) mustSend(m message.Message, recv *eventual2go.Future) {
	recv.Then(c.onRecv(m))
	c.r.Fire(mustSendEvent{}, m)
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
		c.r.Lock()
		defer c.r.Unlock()
		delete(c.mustSendRegister, m)
		return nil
	}
}
