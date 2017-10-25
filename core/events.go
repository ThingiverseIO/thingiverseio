package core

type afterConnectedEvent struct{}

type afterPeerRemovedEvent struct{}

type connectEvent struct{}

type endEvent struct{}

type leaveEvent struct{}

type requestEvent struct{}

type replyEvent struct{}

type startListenEvent struct{}

type stopListenEvent struct{}

type startObserveEvent struct{}

type stopObserveEvent struct{}

type getPropertyEvent struct {}

type setPropertyEvent struct {
	name string
}

type startConsumeEvent struct{}

type stopConsumeEvent struct{}

type addStreamEvent struct {
	name string
}

type mustSendEvent struct{}
