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
	messageIn          *connection.MessageStreamController
	arrive             *config.UUIDStreamController
	leave              *config.UUIDStreamController
	guaranteedMessages map[messages.Message]*peer.Peer
	logger             *log.Logger
	shutdown           *eventual2go.Shutdown
	shutdownComplete   *eventual2go.Completer
}

func New(cfg *config.Config) (m *Manager, err error) {
	m = &Manager{
		r:                  eventual2go.NewReactor(),
		connected:          typed_events.NewBoolStreamController(),
		peers:              map[config.UUID]*peer.Peer{},
		arrive:             config.NewUUIDStreamController(),
		leave:              config.NewUUIDStreamController(),
		guaranteedMessages: map[messages.Message]*peer.Peer{},
		messageIn:          connection.NewMessageStreamController(),
		logger:             log.New(cfg.Logger(), fmt.Sprintf("%s MANAGER ", cfg.UUID()), 0),
		shutdown:           eventual2go.NewShutdown(),
		shutdownComplete:   eventual2go.NewCompleter(),
	}

	m.r.React(peerLeave, m.peerLeave)
	m.r.React(peerConnected, m.peerConnected)
	m.r.OnShutdown(m.onShutdown)

	for _, iface := range cfg.Interfaces() {

		var i *connection.Incoming
		i, err = connection.NewIncoming(iface)
		if err != nil {
			m.logger.Println("ERROR setting up incoming:", err)
			m.Shutdown()
			return
		}
		m.shutdown.Register(i)
		m.messageIn.Join(i.In().Where(filterMessages))

		var t *tracker.Tracker
		t, err = tracker.New(iface, i.Port(), cfg)
		if err != nil {
			m.logger.Println("ERROR setting up tracker:", err)
			m.Shutdown()
			return
		}
		m.tracker = append(m.tracker, t)
		m.shutdown.Register(t)
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

func (m *Manager) PeerArrive() *config.UUIDStream {
	return m.leave.Stream()
}

func (m *Manager) PeerLeave(uuid config.UUID) *config.UUIDFuture {
	return m.leave.Stream().FirstWhere(isPeer(uuid))
}

func isPeer(uuid config.UUID) config.UUIDFilter {
	return func(id config.UUID) bool {
		return id == uuid
	}
}

func (m *Manager) Shutdown() {
	m.r.Shutdown(nil)
	m.shutdownComplete.Future().WaitUntilComplete()
}

func (m *Manager) onShutdown(eventual2go.Data) {
	m.logger.Println("shutting down")
	errs := m.shutdown.Do(nil)
	for _, err := range errs {
		m.logger.Println("ERROR shutdown", err)
	}
	m.logger.Println("shutdown complete")
	m.shutdownComplete.Complete(errs)
}

func (m *Manager) Connected() *typed_events.BoolStream {
	return m.connected.Stream()
}

func (m *Manager) Messages() *connection.MessageStream {
	return m.messageIn.Stream()
}

func (m *Manager) MessagesOfType(t messages.MessageType) *connection.MessageStream {
	return m.messageIn.Stream().Where(filterMsgType(t))
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
	if len(m.peers) == 0 {
		m.logger.Println("no peer found for guaranteed message, storing")
		m.guaranteedMessages[msg] = nil
	} else {
		for _, p := range m.peers {
			m.logger.Println("Sending guaranteed message to", p.UUID())
			m.guaranteedMessages[msg] = p
			p.Send(msg)
			return
		}
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

	m.arrive.Add(p.UUID())
	if len(m.peers) == 0 {
		m.connected.Add(true)
	}

	for msg, p := range m.guaranteedMessages {
		if p == nil {
			go m.SendGuaranteed(msg)
		}
	}

	m.peers[p.UUID()] = p
}

func (m *Manager) peerLeave(d eventual2go.Data) {
	l := d.(hasUuid)

	if p, ok := m.peers[l.UUID()]; ok {
		m.logger.Println("Peer left", l.UUID())
		for msg, r := range m.guaranteedMessages {
			if r == p {
				m.logger.Println("Peer had guarantteed message, resending")
				go m.SendGuaranteed(msg)
			}
		}
		p.Remove()
		delete(m.peers, l.UUID())
		if len(m.peers) == 0 {
			m.connected.Add(false)
		}
		m.leave.Add(l.UUID())
	}
}

func filterMessages(msg connection.Message) bool {
	return messages.PeakType(msg.Payload) != messages.HELLO && messages.PeakType(msg.Payload) != messages.HELLOOK && messages.PeakType(msg.Payload) != messages.DOHAVE && messages.PeakType(msg.Payload) != messages.HAVE && messages.PeakType(msg.Payload) != messages.CONNECT && messages.PeakType(msg.Payload) != messages.END
}

func filterMsgType(t messages.MessageType) connection.MessageFilter {
	return func(m connection.Message) bool {
		return messages.PeakType(m.Payload) == t
	}
}
