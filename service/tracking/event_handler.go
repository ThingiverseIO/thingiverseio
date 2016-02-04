package tracking

import (
	"github.com/hashicorp/memberlist"
	"github.com/joernweissenborn/eventual2go"
)

func newEventHandler() (eh eventHandler) {
	eh.join = eventual2go.NewStreamController()
	eh.leave = eventual2go.NewStreamController()
	return
}

type eventHandler struct {
	join  *eventual2go.StreamController
	leave *eventual2go.StreamController
}

func (eh eventHandler) Join() *eventual2go.Stream {
	return eh.join.Stream
}

func (eh eventHandler) NotifyJoin(n *memberlist.Node) {
	eh.join.Add(n)
}

func (eh eventHandler) Leave() *eventual2go.Stream {
	return eh.leave.Stream
}

func (eh eventHandler) NotifyLeave(n *memberlist.Node) {
	eh.leave.Add(n)
}

func (eh eventHandler) NotifyUpdate(n *memberlist.Node) {
	// not handled at the moment
}
