package manager

import "github.com/joernweissenborn/thingiverseio/config"

type hasUuid interface {
	UUID() config.UUID
}
