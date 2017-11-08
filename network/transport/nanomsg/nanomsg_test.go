package nanomsg_test

import (
	"testing"

	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/network/transport/nanomsg"
)

func TestNanoMessage(t *testing.T) {
	transport1 := &nanomsg.Transport{}
	transport2 := &nanomsg.Transport{}

	network.TransportTestsuite(transport1, transport2, t)
}
