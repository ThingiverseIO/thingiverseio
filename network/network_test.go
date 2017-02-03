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
