package network

import (
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/uuid"
	"github.com/joernweissenborn/eventual2go"
)

//go:generate event_generator -t Message

type Message struct {
	Sender  uuid.UUID
	Type    message.Type
	Payload [][]byte
}

func (m Message) FromSender(sender uuid.UUID) (is bool) {
	is = m.Sender == sender
	return
}

func (m Message) OfType(t message.Type) (is bool) {
	is = m.Type == t
	return
}

func (m Message) Decode() (msg message.Message) {
	msg = message.GetByType(m.Type)
	msg.Unflatten(m.Payload)
	return
}

func FromSender(sender uuid.UUID) MessageFilter {
	return func(m Message) (is bool) {
		return m.FromSender(sender)
	}
}

func OfType(t message.Type) MessageFilter {
	return func(m Message) (has bool) {
		has = m.OfType(t)
		return
	}
}

func ToMessage(d eventual2go.Data) eventual2go.Data {
	return d.(Message).Decode()
}
