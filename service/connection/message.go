package connection

import "github.com/joernweissenborn/eventual2go"


//go:generate event_generator -t Message

type Message struct {
	Iface   string
	Sender  string
	Payload []string
}

func IsMsgFromSender(sender string) MessageFilter {
	return func(m Message) bool {
		return sender == m.Sender
	}
}

type outgoingMessage struct {
	sent *eventual2go.Completer
	payload [][]byte
}
