package tracker

import (
	"sync"

	"github.com/hashicorp/memberlist"
)

func newEventHandler() (eh *eventHandler) {
	eh = &eventHandler{
		Mutex: &sync.Mutex{},
		join:  NewNodeStreamController(),
		leave: NewNodeStreamController(),
	}
	return
}

type eventHandler struct {
	*sync.Mutex
	join  *NodeStreamController
	leave *NodeStreamController
}

func (eh *eventHandler) Join() *NodeStream {
	return eh.join.Stream()
}

func (eh *eventHandler) NotifyJoin(n *memberlist.Node) {
	eh.Lock()
	defer eh.Unlock()
	eh.join.Add(Node{n})
}

func (eh *eventHandler) Leave() *NodeStream {
	return eh.leave.Stream()
}

func (eh *eventHandler) NotifyLeave(n *memberlist.Node) {
	eh.Lock()
	defer eh.Unlock()
	eh.leave.Add(Node{n})
}

func (eh *eventHandler) NotifyUpdate(n *memberlist.Node) {
	// not handled at the moment
}

func (eh *eventHandler) close() {
	eh.Lock()
	defer eh.Unlock()
}
