package network

import (
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/uuid"
	"github.com/joernweissenborn/eventual2go"
)

type Connection struct {
	eventual2go.ActorMessageStream
	UUID uuid.UUID
}

func (c Connection) Send(msg message.Message) {
	c.ActorMessageStream.Send(Message{
		Payload: msg.Flatten(),
		Type:    msg.GetType(),
	})
}

func (c Connection) Close() {
	c.Shutdown(nil)
}
