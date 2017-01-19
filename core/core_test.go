package core_test

import (
	"testing"
	"time"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/core"
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ThingiverseIO/thingiverseio/network"
)

var Descriptor1 = "tag TEST:1"

func TestBasicConnection(t *testing.T) {

	desc, _ := descriptor.Parse(Descriptor1)
	cfg := config.DefaultLocalhost()

	mt1 := &network.MockTracker{}
	mt2 := &network.MockTracker{}
	mps := network.NewMockProvider(2)

	i, _ := core.NewInputCore(desc, cfg, mt1, mps[0])
	o, _ := core.NewOutputCore(desc, cfg, mt2, mps[1])

	arr := network.Arrival{
		IsOutput: true,
		Details:  mt2.Dt,
		UUID:     o.UUID(),
	}

	mt1.Av.Add(arr)

	if !i.ConnectedFuture().WaitUntilTimeout(100*time.Millisecond) ||
		!o.ConnectedFuture().WaitUntilTimeout(100*time.Millisecond) {
		t.Fatal("Peers didn't connect.")
	}

}
