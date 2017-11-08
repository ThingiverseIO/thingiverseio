package network

import (
	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/uuid"
	"github.com/joernweissenborn/eventual2go"
)

type Tracker interface {
	eventual2go.Shutdowner

	// Init initializes the tracker.
	Init(cfg *config.Config, details [][]byte) error

	// Arrivals return a stream of arrived peers.
	Arrivals() *ArrivalStream

	// Leaving returns a stream of leaving peers.
	Leaving() *uuid.UUIDStream

	// StartAdvertisment starts advertisment.
	StartAdvertisment() error

	// StopAdvertisment stops advertisment.
	StopAdvertisment()
}
