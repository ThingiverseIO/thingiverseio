package network

import "testing"

func TestProviderMock(t *testing.T) {
	provs := NewMockProvider(2)

	ProviderTestsuite(provs[0], provs[1], t)
}

func TestTrackerMock(t *testing.T) {
	arr1 := NewArrivalStreamController()
	arr2 := NewArrivalStreamController()
	t1 := &MockTracker{Av: arr1, Partner: arr2}
	t2 := &MockTracker{Av: arr2, Partner: arr1}
	TrackerTestSuite(t1, t2, t)
}
