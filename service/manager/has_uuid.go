package manager

import "github.com/ThingiverseIO/thingiverseio/config"

type hasUuid interface {
	UUID() config.UUID
}
