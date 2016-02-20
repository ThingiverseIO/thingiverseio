package tracker

import "github.com/hashicorp/memberlist"

func newEventHandler() (eh eventHandler) {
	eh.join = NewNodeStreamController()
	eh.leave = NewNodeStreamController()
	return
}

type eventHandler struct {
	join  *NodeStreamController
	leave *NodeStreamController
}

func (eh eventHandler) Join() *NodeStream {
	return eh.join.Stream()
}

func (eh eventHandler) NotifyJoin(n *memberlist.Node) {
	eh.join.Add(Node{n})
}

func (eh eventHandler) Leave() *NodeStream {
	return eh.leave.Stream()
}

func (eh eventHandler) NotifyLeave(n *memberlist.Node) {
	eh.leave.Add(Node{n})
}

func (eh eventHandler) NotifyUpdate(n *memberlist.Node) {
	// not handled at the moment
}
