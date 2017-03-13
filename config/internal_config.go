package config

import (
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ThingiverseIO/thingiverseio/uuid"
)

type InternalConfig struct {
	Output bool
	Tags   descriptor.Tagset
	UUID   uuid.UUID
}

func NewInternalConfig(output bool, tags descriptor.Tagset) *InternalConfig {
	return &InternalConfig{
		Output: output,
		Tags:   tags,
		UUID:   uuid.New(),
	}
}
