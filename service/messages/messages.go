package messages

//go:generate stringer -type=MessageType

// MessageType represents a message constant.
type MessageType int

// Message type constants
const (
	HELLO MessageType = iota
	HELLOOK
	DOHAVE
	HAVE
	CONNECT
	REQUEST
	RESULT
	LISTEN
	STOPLISTEN
	END
	REJECT
	MOCK
)

var msgs = map[MessageType]Message{}

func registerMessage(m Message){
	msgs[m.GetType()] = m
}

func Get(messagetype MessageType) (msg Message) {
	msg = msgs[messagetype]
	/*
	switch messagetype {
	case HELLO:
		msg = new(Hello)
	case HELLOOK:
		msg = new(HelloOk)
	case DOHAVE:
		msg = new(DoHave)
	case HAVE:
		msg = new(Have)
	case CONNECT:
		msg = new(Connect)
	case REQUEST:
		msg = new(Request)
	case RESULT:
		msg = new(Result)
	case LISTEN:
		msg = new(Listen)
	case STOPLISTEN:
		msg = new(StopListen)
	case END:
		msg = new(End)
	case MOCK:
		msg = new(Mock)
	}
	*/
	return
}

func Is(t MessageType) MessageFilter {
	return func(d Message) bool {
		return d.GetType() == t
	}
}
