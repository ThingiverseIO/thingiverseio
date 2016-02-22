package manager

import "github.com/joernweissenborn/thingiverse.io/config"

type hasUuid interface {
	UUID() config.UUID
}
