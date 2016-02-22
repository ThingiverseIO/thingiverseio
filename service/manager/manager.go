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
	peerSend      = "peer_send"
)

type msgToSend struct {
	uuid config.UUID
	m    messages.Message
}

type Manager struct {
	r *eventual2go.Reactor

	connected *typed_events.BoolStreamController
	peers     map[config.UUID]*peer.Peer
	tracker   []*tracker.Tracker

	logger *log.Logger
}

func New(cfg *config.Config) (m *Manager, err error) {
	m = &Manager{
		r:         eventual2go.NewReactor(),
		connected: typed_events.NewBoolStreamController(),
		peers:     map[config.UUID]*peer.Peer{},
		logger:    log.New(cfg.Logger(), fmt.Sprintf("%s MANAGER ", cfg.UUID()), 0),
	}

	m.r.React(peerLeave, m.peerLeave)
	m.r.React(peerConnected, m.peerConnected)
	m.r.React(peerSend, m.peerSend)

	for _, iface := range cfg.Interfaces() {

		var i *connection.Incoming
		i, err = connection.NewIncoming(iface)
		if err != nil {
			return
		}

		var t *tracker.Tracker
		t, err = tracker.New(iface, i.Port(), cfg)
		if err != nil {
			return
		}
		m.tracker = append(m.tracker, t)
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

func (m *Manager) Connected() *typed_events.BoolStream {
	return m.connected.Stream()
}

func (m *Manager) SendTo(uuid config.UUID, msg messages.Message) {
	m.r.Fire(peerSend, &msgToSend{uuid, msg})
}

func (m *Manager) SendAll(msg messages.Message) {
	m.r.Fire(peerSend, &msgToSend{"", msg})
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
		p.Remove()
		delete(m.peers, l.UUID())
		if len(m.peers) == 0 {
			m.connected.Add(false)
		}
	}

}

func (m *Manager) peerSend(d eventual2go.Data) {
	msg := d.(*msgToSend)

	if msg.uuid == "" {
		for _, p := range m.peers {
			p.Send(msg.m)
		}
	} else {
		if p, ok := m.peers[msg.uuid]; ok {
			p.Send(msg.m)
		}
	}
}
