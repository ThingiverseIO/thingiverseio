package memberlist_test

import (
	"testing"

	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/network/tracker/memberlist"
)

func TestMemberlistTracker(t *testing.T) {
	tracker1 := &memberlist.Tracker{}
	tracker2 := &memberlist.Tracker{}

	network.TrackerTestSuite(tracker1, tracker2, t)
}
