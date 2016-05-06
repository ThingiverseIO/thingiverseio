package discoverer

import (
	"fmt"
	"log"
	"sync"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/service/connection"
	"github.com/ThingiverseIO/thingiverseio/service/messages"
	"github.com/ThingiverseIO/thingiverseio/service/peer"
	"github.com/ThingiverseIO/thingiverseio/service/tracker"
)

// Discoverer is a reactor for service discovery.
type Discoverer struct {
	cfg            *config.Config
	incoming       *connection.Incoming
	connectedPeers *peer.PeerStreamController
	seenNodes      map[config.UUID]*peer.Peer
	m              *sync.Mutex
	logger         *log.Logger
}

// New Creates a new Discoverer.
func New(join *tracker.NodeStream, incoming *connection.Incoming, cfg *config.Config) (d *Discoverer) {
	d = &Discoverer{
		cfg:            cfg,
		incoming:       incoming,
		connectedPeers: peer.NewPeerStreamController(),
		seenNodes:      map[config.UUID]*peer.Peer{},
		m:              &sync.Mutex{},
		logger:         log.New(cfg.Logger(), fmt.Sprintf("%s DISCOVERER ", cfg.UUID()), 0),
	}

	//TODO replace with listenwhere
	join.Where(d.isNodeInteresting).Listen(d.onInterestingNode)
	incoming.Messages().Where(messages.Is(messages.HELLO)).Listen(d.onHello)
	return
}

func (d *Discoverer) ConnectedPeers() *peer.PeerStream {
	return d.connectedPeers.Stream()
}

func (d *Discoverer) nodeSeen(uuid config.UUID) (p *peer.Peer, s bool) {
	p, s = d.seenNodes[uuid]
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
	d.m.Lock()
	defer d.m.Unlock()
	h := m.(*messages.Hello)
	d.logger.Println("Got HELLO from", h.UUID)
	if p, s := d.nodeSeen(config.UUID(h.UUID)); s {
		d.logger.Println("Node is already known")
		p.Send(&messages.HelloOk{})
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
	d.seenNodes[p.UUID()] = p
}

func (d *Discoverer) onInterestingNode(node tracker.Node) {
	d.m.Lock()
	defer d.m.Unlock()
	meta, _ := node.Meta()
	uuid := node.UUID()
	d.logger.Println("Found interesting node", string(uuid))
	if p, s := d.nodeSeen(uuid); s {
		p.Send(&messages.HelloOk{})
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
	d.seenNodes[p.UUID()] = p

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
