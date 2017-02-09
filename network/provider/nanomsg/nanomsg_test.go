package nanomsg_test

import (
	"testing"

	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/network/provider/nanomsg"
)

func TestNanoMessage(t *testing.T) {
	provider1 := &nanomsg.Provider{}
	provider2 := &nanomsg.Provider{}

	network.ProviderTestsuite(provider1, provider2, t)
}
