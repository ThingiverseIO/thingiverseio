package nanomsg_test

import (
	"testing"

	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/network/transport/nanomsg"
)

func TestNanoMessage(t *testing.T) {
	transport1 := &nanomsg.Provider{}
	transport2 := &nanomsg.Provider{}

	network.ProviderTestsuite(transport1, transport2, t)
}
