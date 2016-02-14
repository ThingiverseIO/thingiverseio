package messages

import "github.com/joernweissenborn/eventual2go"

//go:generate stringer -type=MessageType

type MessageType int

const (
	HELLO MessageType = iota
	HELLO_OK
	DO_HAVE
	HAVE
	REQUEST
	RESULT
	LISTEN
	STOP_LISTEN
)

func Get(messagetype MessageType) (msg Message) {

	switch messagetype {
	case HELLO:
		msg = new(Hello)
	case HELLO_OK:
		msg = new(HelloOk)
	case REQUEST:
		msg = new(Request)
	case RESULT:
		msg = new(Result)
	case LISTEN:
		msg = new(Listen)
	case STOP_LISTEN:
		msg = new(StopListen)
	}
	return
}

func Is(t MessageType) eventual2go.Filter {
	return func(d eventual2go.Data) bool {
		return d.(Message).GetType() == t
	}
}
