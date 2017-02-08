package network

import (
	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/uuid"
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

	// Run starts the tracker.
	Run()
}
