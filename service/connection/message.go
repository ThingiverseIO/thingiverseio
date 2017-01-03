package connection

import (
	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/service/messages"
	"github.com/joernweissenborn/eventual2go"
)

//go:generate event_generator -t Message

type Message struct {
	Iface   string
	Sender  config.UUID
	Type    messages.MessageType
	Payload []byte
}

func isMsgFromSender(sender config.UUID) MessageFilter {
	return func(m Message) bool {
		return sender == m.Sender
	}
}

func ToMessage(d eventual2go.Data) eventual2go.Data {
	m := d.(messages.FlatMessage)
	return messages.Unflatten(m)
}

type outgoingMessage struct {
	sent    *eventual2go.Completer
	payload [][]byte
}
