package core

import (
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/network/transport/nanomsg"
	"github.com/ThingiverseIO/thingiverseio/network/tracker/memberlist"
)

func DefaultBackends() (tracker network.Tracker, transport []network.Transport) {
	tracker = &memberlist.Tracker{}
	transport = []network.Transport{&nanomsg.Transport{}}
	return
}
