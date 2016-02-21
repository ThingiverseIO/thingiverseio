package discoverer

import (
	"log"
	"strings"
	"sync"

	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/joernweissenborn/thingiverse.io/service/connection"
	"github.com/joernweissenborn/thingiverse.io/service/messages"
	"github.com/joernweissenborn/thingiverse.io/service/peer"
	"github.com/joernweissenborn/thingiverse.io/service/tracker"
)

// Discoverer is a reactor for service discovery.
type Discoverer struct {
	cfg *config.Config

	incoming *connection.Incoming

	connectedPeers *peer.PeerStreamController

	seenNodes map[string]interface{}
	m         *sync.Mutex

	logger *log.Logger
}

// New Creates a new Discoverer.
func New(join *tracker.NodeStream, incoming *connection.Incoming, cfg *config.Config) (d *Discoverer) {
	d = &Discoverer{
		cfg:            cfg,
		incoming:       incoming,
		connectedPeers: peer.NewPeerStreamController(),
		seenNodes:      map[string]interface{}{},
		m:              &sync.Mutex{},
		logger:         log.New(cfg.Logger(), "DISCOVERER ", 0),
	}

	join.Where(d.isNodeInteresting).Listen(d.onInterestingNode)
	incoming.Messages().Where(messages.Is(messages.HELLO)).Listen(d.onHello)
	return
}

func (d *Discoverer) ConnectedPeers() *peer.PeerStream {
	return d.connectedPeers.Stream()
}

func (d *Discoverer) nodeSeen(uuid string) (s bool) {
	d.m.Lock()
	defer d.m.Unlock()
	_, s = d.seenNodes[uuid]
	if !s {
		d.seenNodes[uuid] = nil
	}
	return
}

func (d *Discoverer) removeFromSeen(p *peer.Peer) *peer.Peer {
	d.m.Lock()
	defer d.m.Unlock()
	_, s := d.seenNodes[p.UUID()]
	if s {
		delete(d.seenNodes, p.UUID())
	}
	return nil
}

func (d *Discoverer) onHello(m messages.Message) {
	h := m.(*messages.Hello)
	d.logger.Println("Got HELLO from", h.UUID)
	if d.nodeSeen(h.UUID) {
		d.logger.Println("Node is already known")
		return
	}
	p, err := peer.NewFromHello(h, d.incoming, d.cfg)
	if err != nil {
		d.logger.Println("ERROR: ", err)
	}
	if !d.cfg.Exporting() {
		p.Check()
	}
	p.Removed().Then(d.removeFromSeen)
	p.Connected().Then(d.removeFromSeen)

	d.connectedPeers.JoinFuture(p.Connected())
}

func (d *Discoverer) onInterestingNode(node tracker.Node) {
	meta, _ := node.Meta()
	uuid := strings.Split(node.Name, ":")[0]
	d.logger.Println("Found interesting node", uuid)
	if d.nodeSeen(uuid) {
		d.logger.Println("Node is already known")
		return
	}

	p, err := peer.New(uuid, node.Addr.String(), meta.Adport, d.incoming, d.cfg)
	if err != nil {
		d.logger.Println("ERROR: ", err)
	}
	p.InitConnection()

	if !d.cfg.Exporting() {
		p.Check()
	}

	d.connectedPeers.JoinFuture(p.Connected())

}

func (d *Discoverer) isNodeInteresting(node tracker.Node) bool {
	meta, err := node.Meta()
	if err != nil {
		return false
	}

	if meta.Exporting == d.cfg.Exporting() {
		return false
	}

	tk, tv, err := meta.TagKeyValue()

	if err != nil {
		return false
	}

	t, f := d.cfg.Tags()[tk]
	if !f {
		return false
	}

	return t == tv
}
