package message

//go:generate event_generator -t Message

//TODO add http://www.ugorji.net/blog/go-codecgen

type Message interface {
	New() Message
	GetType() Type
	Flatten() [][]byte
	Unflatten([][]byte)
}

var msgs = map[Type]Message{}

func registerMessage(m Message) {
	msgs[m.GetType()] = m
}

func GetByType(messagetype Type) (msg Message) {
	msg = msgs[messagetype].New()
	return
}
