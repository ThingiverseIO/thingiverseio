package peer

import (
	"testing"
	"time"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/service/connection"
	"github.com/ThingiverseIO/thingiverseio/service/messages"
)

func TestInitConnection(t *testing.T) {

	i1, err := connection.NewIncoming("127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	i2, err := connection.NewIncoming("127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	cfg1 := config.New(true, map[string]string{})
	cfg2 := config.New(false, map[string]string{})

	c := i2.MessagesFromSender(cfg1.UUID()).Where(messages.Is(messages.HELLO)).AsChan()

	p1, err := New(cfg2.UUID(), "127.0.0.1", i2.Port(), i1, cfg1)
	if err != nil {
		t.Fatal(err)
	}
	p1.InitConnection()

	var p2 *Peer
	select {
	case <-time.After(5 * time.Second):
		t.Fatal("Did not received Hello")
	case d := <-c:
		m := d.(*messages.Hello)
		if m.Address != i1.Addr() || m.Port != i1.Port() {
			t.Fatal("Wrong message", m)
		}
		p2, err = NewFromHello(m, i2, cfg2)
		if err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(5 * time.Second)

	if !p1.initialized.Completed() {
		t.Error("Connection 1 did not initialize")
	}
	if !p2.initialized.Completed() {
		t.Error("Connection 2 did not initialize")
	}

}

func TestConnecting(t *testing.T) {

	i1, err := connection.NewIncoming("127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	i2, err := connection.NewIncoming("127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	cfg1 := config.New(true, map[string]string{})
	cfg1.AddOrSetUserTag("tag1", "1")
	cfg1.AddOrSetUserTag("tag2", "2")
	cfg2 := config.New(false, map[string]string{})
	cfg2.AddOrSetUserTag("tag2", "2")

	p1, err := New(cfg2.UUID(), "127.0.0.1", i2.Port(), i1, cfg1)
	if err != nil {
		t.Fatal(err)
	}

	p2, err := NewFromHello(&messages.Hello{string(cfg1.UUID()), "127.0.0.1", i1.Port()}, i2, cfg2)
	if err != nil {
		t.Fatal(err)
	}
	p2.Check()
	time.Sleep(5 * time.Second)

	if !p1.Connected().Completed() {
		t.Error("Connection 1 did not initialize")
	}
	if !p2.Connected().Completed() {
		t.Error("Connection 2 did not initialize")
	}

}

func TestNotConnecting(t *testing.T) {

	i1, err := connection.NewIncoming("127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	i2, err := connection.NewIncoming("127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	cfg1 := config.New(true, map[string]string{})
	cfg1.AddOrSetUserTag("tag1", "1")
	cfg1.AddOrSetUserTag("tag2", "2")
	cfg2 := config.New(false, map[string]string{})
	cfg2.AddOrSetUserTag("tag2", "2")
	cfg2.AddOrSetUserTag("tag1", "2")

	p1, err := New(cfg2.UUID(), "127.0.0.1", i2.Port(), i1, cfg1)
	if err != nil {
		t.Fatal(err)
	}

	p2, err := NewFromHello(&messages.Hello{string(cfg1.UUID()), "127.0.0.1", i1.Port()}, i2, cfg2)
	if err != nil {
		t.Fatal(err)
	}
	p2.Check()
	time.Sleep(5 * time.Second)

	if p1.Connected().Completed() {
		t.Error("Connection 1 did initialize")
	}
	if p2.Connected().Completed() {
		t.Error("Connection 2 did initialize")
	}

}
