package discoverer

import (
	"github.com/hashicorp/memberlist"
	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/joernweissenborn/thingiverse.io/service/connection"
)

type Discoverer struct {
	cfg   *config.Config
	msgIn *eventual2go.Stream
}

func New(peers *eventual2go.Stream, cfg *config.Config) (d *Discoverer) {
	d = &Discoverer{}

	return
}

func (d *Discoverer) newPeer(node memberlist.Node) {
	peerIn := d.msgIn.Where(connection.IsMsgFromSender(node.Name))

}
