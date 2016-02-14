package connection

import "github.com/joernweissenborn/eventual2go"

type Message struct {
	Iface   string
	Sender  string
	Payload []string
}

func IsMsgFromSender(sender string) eventual2go.Filter {
	return func(d eventual2go.Data) bool {
		m := d.(Message)
		return sender == m.Sender
	}
}

type outgoingMessage struct {
	sent *eventual2go.Completer
	payload [][]byte
}
