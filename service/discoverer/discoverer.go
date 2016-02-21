package discoverer

import (
	"log"

	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/joernweissenborn/thingiverse.io/service/connection"
	"github.com/joernweissenborn/thingiverse.io/service/tracker"
	"github.com/joernweissenborn/thingiverse.io/service/peer"
)

// Discoverer is a reactor for service discovery.
type Discoverer struct {
	cfg *config.Config

	incoming *connection.Incoming

	peers map[string]*peer.Peer

	logger *log.Logger
}

// New Creates a new Discoverer.
func New(peers *eventual2go.Stream, cfg *config.Config) (d *Discoverer) {
	d = &Discoverer{
		cfg:    cfg,
		logger: log.New(cfg.Logger(), "DISCOVERER ", 0),
	}

	return
}

func (d *Discoverer) onInterestingNode(n eventual2go.Data) {
	node := n.(tracker.Node)
	meta, _ := node.Meta()
	d.logger.Println("Found interesting node", node.Name)
	if _, f := d.peers[node.Name]; f {
		d.logger.Println("Node is already known")
		return
	}

	p, err := peer.New(node.Name, node.Addr.String(), meta.Adport, d.incoming, d.cfg)
	if err != nil {
		d.logger.Println("ERROR: ", err)
	}
	p.InitConnection()

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
