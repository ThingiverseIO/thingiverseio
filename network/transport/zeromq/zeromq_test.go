package zeromq_test

import (
	"testing"

	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/network/provider/zeromq"
)

func TestNanoMessage(t *testing.T) {
	provider1 := &zeromq.Provider{}
	provider2 := &zeromq.Provider{}

	network.ProviderTestsuite(provider1, provider2, t)
}
