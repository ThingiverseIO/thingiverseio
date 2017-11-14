package network

import (
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/uuid"
	"github.com/joernweissenborn/eventual2go"
)

//go:generate event_generator -t Package

type Package struct {
	Sender  uuid.UUID
	Type    message.Type
	Payload [][]byte
}

func (p Package) FromSender(sender uuid.UUID) (is bool) {
	is = p.Sender == sender
	return
}

func (p Package) OfType(t message.Type) (is bool) {
	is = p.Type == t
	return
}

func (p Package) Decode() (msg message.Message) {
	msg = message.GetByType(p.Type)
	msg.Unflatten(p.Payload)
	return
}

func FromSender(sender uuid.UUID) PackageFilter {
	return func(p Package) (is bool) {
		return p.FromSender(sender)
	}
}

func OfType(t message.Type) PackageFilter {
	return func(p Package) (has bool) {
		has = p.OfType(t)
		return
	}
}

func ToMessage(d eventual2go.Data) eventual2go.Data {
	return d.(Package).Decode()
}
