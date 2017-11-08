package zeromq_test

import (
	"testing"

	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/network/transport/zeromq"
)

func TestNanoMessage(t *testing.T) {
	transport1 := &zeromq.Transport{}
	transport2 := &zeromq.Transport{}

	network.TransportTestsuite(transport1, transport2, t)
}
