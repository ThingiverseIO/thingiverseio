package network

import (
	"bytes"
	"testing"
	"time"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/uuid"
)

func testMsg() *message.Mock {
	return &message.Mock{
		Data: [][]byte{[]byte{1, 2, 3}, []byte{4, 5, 6}},
	}
}

func checkMessage(m Message, sender uuid.UUID, t *testing.T) {
	if m.Sender != sender {
		t.Error("Wrond sender", sender, m.Sender)
	}

	if m.Type != message.MOCK {
		t.Error("Wrong type", m.Type)
	}

	if !bytes.Equal(m.Payload[0], testMsg().Data[0]) || !bytes.Equal(m.Payload[1], testMsg().Data[1]) {
		t.Error("Wrong payload", m.Payload, testMsg().Data)
	}

}

func ProviderTestsuite(provider1, provider2 Provider, t *testing.T) {

	uuid1 := uuid.New()
	cfg1 := &config.Config{
		Internal: &config.InternalConfig{UUID: uuid1},
		User:     config.DefaultLocalhost(),
	}
	if err := provider1.Init(cfg1); err != nil {
		t.Fatal("Error on initialzing provider1", err)
	}

	uuid2 := uuid.New()
	cfg2 := &config.Config{
		Internal: &config.InternalConfig{UUID: uuid2},
		User:     config.DefaultLocalhost(),
	}

	if err := provider2.Init(cfg2); err != nil {
		t.Fatal("Error on initialzing provider2", err)
	}

	conn1, err := provider1.Connect(provider2.Details(), uuid2)
	if err != nil {
		t.Fatal("Error on connecting to provider2", err)
	}

	conn2, err := provider2.Connect(provider1.Details(), uuid2)
	if err != nil {
		t.Fatal("Error on connecting to provider1", err)
	}

	msg1 := provider1.Messages().First()
	msg2 := provider2.Messages().First()

	conn1.Send(testMsg())
	conn2.Send(testMsg())

	if !msg1.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Message 1 did not arrive.")
	}

	if !msg2.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Message 2 did not arrive.")
	}

	checkMessage(msg1.Result(), uuid2, t)
	checkMessage(msg2.Result(), uuid1, t)

}

func TestProviderMock(t *testing.T) {
	provs := NewMockProvider(2)

	ProviderTestsuite(provs[0], provs[1], t)
}

func TrackerTestSuite(tracker1, tracker2 Tracker, t *testing.T) {

	uuid1 := uuid.New()
	cfg1 := &config.Config{
		Internal: &config.InternalConfig{UUID: uuid1},
		User:     config.DefaultLocalhost(),
	}

	details1 := [][]byte{[]byte{1, 2, 3}}

	if err := tracker1.Init(cfg1, details1); err != nil {
		t.Fatal("Error on initialzing provider1", err)
	}

	arr1 := tracker1.Arrivals().First()

	uuid2 := uuid.New()
	cfg2 := &config.Config{
		Internal: &config.InternalConfig{UUID: uuid2},
		User:     config.DefaultLocalhost(),
	}

	details2 := [][]byte{[]byte{4, 5, 6}}

	if err := tracker2.Init(cfg2, details2); err != nil {
		t.Fatal("Error on initialzing provider2", err)
	}

	arr2 := tracker2.Arrivals().First()

	tracker1.Run()
	tracker2.Run()

	if !arr1.WaitUntilTimeout(1*time.Second) || !arr2.WaitUntilTimeout(1*time.Second) {
		t.Fatal("Trackers did not find each other.")
	}

	if arr1.Result().UUID != uuid2 {
		t.Error("Arrival 1 has wrong uuid", arr1.Result().UUID, uuid2)
	}

	if arr2.Result().UUID != uuid1 {
		t.Error("Arrival 2 has wrong uuid", arr2.Result().UUID, uuid1)
	}

	if !bytes.Equal(arr1.Result().Details[0], details2[0]) {
		t.Error("Arrival 1 has wrong details", arr1.Result().Details, details2)
	}

	if !bytes.Equal(arr2.Result().Details[0], details1[0]) {
		t.Error("Arrival 1 has wrong details", arr2.Result().Details, details1)
	}
}

func TestTrackerMock(t *testing.T) {
	arr1 := NewArrivalStreamController()
	arr2 := NewArrivalStreamController()
	t1 := &MockTracker{Av: arr1, Partner: arr2}
	t2 := &MockTracker{Av: arr2, Partner: arr1}
	TrackerTestSuite(t1, t2, t)
}
