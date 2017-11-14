package network

import (
	"bytes"
	"testing"
	"time"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/uuid"
)

func testMsg() *message.Mock {
	return &message.Mock{
		Data: [][]byte{[]byte{1, 2, 3}, []byte{4, 5, 6}},
	}
}

func checkPackage(p Package, sender uuid.UUID, t *testing.T) {
	if p.Sender != sender {
		t.Error("Wrond sender", sender, p.Sender)
	}

	if p.Type != message.MOCK {
		t.Error("Wrong type", p.Type)
	}

	if !bytes.Equal(p.Payload[0], testMsg().Data[0]) || !bytes.Equal(p.Payload[1], testMsg().Data[1]) {
		t.Error("Wrong payload", p.Payload, testMsg().Data)
	}

}

func TransportTestsuite(transport1, transport2 Transport, t *testing.T) {

	uuid1 := uuid.New()
	cfg1 := &config.Config{
		Internal: &config.InternalConfig{UUID: uuid1},
		User:     config.DefaultLocalhost(),
	}
	if err := transport1.Init(cfg1); err != nil {
		t.Fatal("Error on initialzing transport1", err)
	}

	uuid2 := uuid.New()
	cfg2 := &config.Config{
		Internal: &config.InternalConfig{UUID: uuid2},
		User:     config.DefaultLocalhost(),
	}

	if err := transport2.Init(cfg2); err != nil {
		t.Fatal("Error on initialzing transport2", err)
	}

	conn1, err := transport1.Connect(transport2.Details(), uuid2)
	if err != nil {
		t.Fatal("Error on connecting to transport2", err)
	}

	conn2, err := transport2.Connect(transport1.Details(), uuid2)
	if err != nil {
		t.Fatal("Error on connecting to transport1", err)
	}

	msg1 := transport1.Packages().First()
	msg2 := transport2.Packages().First()

	conn1.Send(testMsg())
	conn2.Send(testMsg())

	if !msg1.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Message 1 did not arrive.")
	}

	if !msg2.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Message 2 did not arrive.")
	}

	checkPackage(msg1.Result(), uuid2, t)
	checkPackage(msg2.Result(), uuid1, t)

}

func TrackerTestSuite(tracker1, tracker2 Tracker, t *testing.T) {

	uuid1 := uuid.New()
	cfg1 := &config.Config{
		Internal: &config.InternalConfig{UUID: uuid1},
		User:     config.DefaultLocalhost(),
	}

	details1 := [][]byte{[]byte{1, 2, 3}}

	if err := tracker1.Init(cfg1, details1); err != nil {
		t.Fatal("Error on initialzing transport1", err)
	}

	arr1 := tracker1.Arrivals().First()

	uuid2 := uuid.New()
	cfg2 := &config.Config{
		Internal: &config.InternalConfig{UUID: uuid2},
		User:     config.DefaultLocalhost(),
	}

	details2 := [][]byte{[]byte{4, 5, 6}}

	if err := tracker2.Init(cfg2, details2); err != nil {
		t.Fatal("Error on initialzing transport2", err)
	}

	arr2 := tracker2.Arrivals().First()

	tracker1.StartAdvertisment()
	tracker2.StartAdvertisment()

	if !arr1.WaitUntilTimeout(10*time.Second) || !arr2.WaitUntilTimeout(10*time.Second) {
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
