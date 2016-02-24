package manager

import (
	"fmt"
	"log"

	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/eventual2go/typed_events"
	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/joernweissenborn/thingiverse.io/service/connection"
	"github.com/joernweissenborn/thingiverse.io/service/discoverer"
	"github.com/joernweissenborn/thingiverse.io/service/messages"
	"github.com/joernweissenborn/thingiverse.io/service/peer"
	"github.com/joernweissenborn/thingiverse.io/service/tracker"
)

const (
	peerLeave     = "peer_leave"
	peerConnected = "peer_connected"
)

type msgToSend struct {
	uuid config.UUID
	m    messages.Message
}

type Manager struct {
	r                  *eventual2go.Reactor
	connected          *typed_events.BoolStreamController
	peers              map[config.UUID]*peer.Peer
	tracker            []*tracker.Tracker
	messageIn          *messages.MessageStreamController
	guaranteedMessages map[messages.Message]*peer.Peer
	logger             *log.Logger
	shutdown           *eventual2go.Completer
	shutdownComplete   *eventual2go.Completer
	shutdownWg         *eventual2go.FutureWaitGroup
}

func New(cfg *config.Config) (m *Manager, err error) {
	m = &Manager{
		r:                  eventual2go.NewReactor(),
		connected:          typed_events.NewBoolStreamController(),
		peers:              map[config.UUID]*peer.Peer{},
		guaranteedMessages: map[messages.Message]*peer.Peer{},
		messageIn:          messages.NewMessageStreamController(),
		logger:             log.New(cfg.Logger(), fmt.Sprintf("%s MANAGER ", cfg.UUID()), 0),
		shutdown:           eventual2go.NewCompleter(),
		shutdownComplete:   eventual2go.NewCompleter(),
		shutdownWg:         eventual2go.NewFutureWaitGroup(),
	}

	m.messageIn.Stream().Listen(func(msg messages.Message) {
		m.logger.Println("Received Massage", msg.GetType())
	})

	m.r.React(peerLeave, m.peerLeave)
	m.r.React(peerConnected, m.peerConnected)
	m.r.OnShutdown(m.onShutdown)

	for _, iface := range cfg.Interfaces() {

		var i *connection.Incoming
		i, err = connection.NewIncoming(iface)
		if err != nil {
			return
		}
		m.shutdown.Future().Then(i.CloseOnFuture)
		m.shutdownWg.Add(i.Closed())
		m.messageIn.Join(i.Messages().Where(filterMessages))

		var t *tracker.Tracker
		t, err = tracker.New(iface, i.Port(), cfg)
		if err != nil {
			return
		}
		m.tracker = append(m.tracker, t)
		m.shutdown.Future().Then(t.StopOnFuture)
		m.shutdownWg.Add(t.Stopped())
		m.r.AddStream(peerLeave, t.Leave().Stream)

		var d *discoverer.Discoverer
		d = discoverer.New(t.Join(), i, cfg)

		m.r.AddStream(peerConnected, d.ConnectedPeers().Stream)
	}

	return
}

func (m *Manager) Run() {
	for _, t := range m.tracker {
		t.StartAutoJoin()
	}
}

func (m *Manager) Shutdown() {
	m.r.Shutdown(nil)
	m.shutdownComplete.Future().WaitUntilComplete()
}

func (m *Manager) onShutdown(eventual2go.Data) {
	m.logger.Println("shutting down")
	m.shutdown.Complete(nil)
	m.shutdownWg.Wait()
	m.shutdownComplete.Complete(nil)
	m.logger.Println("shutdown complete")
}

func (m *Manager) Connected() *typed_events.BoolStream {
	return m.connected.Stream()
}

func (m *Manager) Messages() *messages.MessageStream {
	return m.messageIn.Stream()
}

func (m *Manager) Send(msg messages.Message) {
	m.r.Lock()
	defer m.r.Unlock()
	for _, p := range m.peers {
		p.Send(msg)
		return
	}
}

func (m *Manager) SendGuaranteed(msg messages.Message) (c *eventual2go.Completer) {
	m.r.Lock()
	defer m.r.Unlock()
	for _, p := range m.peers {
		m.guaranteedMessages[msg] = p
		p.Send(msg)
		return
	}
	c = eventual2go.NewCompleter()
	c.Future().Then(m.acknowdledeReception(msg))
	return
}

func (m *Manager) acknowdledeReception(msg messages.Message) eventual2go.CompletionHandler {
	return func(eventual2go.Data) eventual2go.Data {
		m.r.Lock()
		defer m.r.Unlock()
		delete(m.guaranteedMessages, msg)
		return nil
	}
}
func (m *Manager) SendTo(uuid config.UUID, msg messages.Message) {
	m.r.Lock()
	defer m.r.Unlock()
	if p, ok := m.peers[uuid]; ok {
		p.Send(msg)
	}
}

func (m *Manager) SendToAll(msg messages.Message) {
	m.r.Lock()
	defer m.r.Unlock()
	for _, p := range m.peers {
		p.Send(msg)
	}
}

func (m *Manager) peerConnected(d eventual2go.Data) {
	p := d.(*peer.Peer)
	m.logger.Println("Successfully connected to", p.UUID())
	m.r.AddFuture(peerLeave, p.Removed().Future)

	if len(m.peers) == 0 {
		m.connected.Add(true)
	}

	m.peers[p.UUID()] = p
}

func (m *Manager) peerLeave(d eventual2go.Data) {
	l := d.(hasUuid)

	if p, ok := m.peers[l.UUID()]; ok {
		for msg, r := range m.guaranteedMessages {
			if r == p {
				go m.SendGuaranteed(msg)
			}
		}
		p.Remove()
		delete(m.peers, l.UUID())
		if len(m.peers) == 0 {
			m.connected.Add(false)
		}
	}
}

func filterMessages(msg messages.Message) bool {
	return msg.GetType() != messages.HELLO && msg.GetType() != messages.HELLOOK && msg.GetType() != messages.DOHAVE && msg.GetType() != messages.HAVE && msg.GetType() != messages.CONNECT && msg.GetType() != messages.END
}
