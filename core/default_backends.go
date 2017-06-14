package core

import (
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/network/provider/nanomsg"
	"github.com/ThingiverseIO/thingiverseio/network/tracker/memberlist"
)

func DefaultBackends() (tracker network.Tracker, provider []network.Provider) {
	tracker = &memberlist.Tracker{}
	provider = []network.Provider{&nanomsg.Provider{}}
	return
}
