package manager

import (
	"os"
	"testing"
	"time"

	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/joernweissenborn/thingiverse.io/service/connection"
	"github.com/joernweissenborn/thingiverse.io/service/messages"
)

func TestManagerConnection(t *testing.T) {

	cfg1 := config.New(os.Stdout, true)
	cfg1.OverrideUUID("exp")
	cfg1.AddOrSetUserTag("tag1", "1")
	cfg1.AddOrSetUserTag("tag2", "2")
	cfg1.OverrideInterfaces([]string{"127.0.0.1"})

	cfg2 := config.New(os.Stdout, false)
	cfg2.OverrideUUID("imp1")
	cfg2.AddOrSetUserTag("tag2", "2")
	cfg2.OverrideInterfaces([]string{"127.0.0.1"})

	cfg3 := config.New(os.Stdout, false)
	cfg3.OverrideUUID("imp2")
	cfg3.AddOrSetUserTag("tag1", "1")
	cfg3.OverrideInterfaces([]string{"127.0.0.1"})

	m1, err := New(cfg1)
	if err != nil {
		t.Fatal(err)
	}
	defer m1.Shutdown()

	m2, err := New(cfg2)
	if err != nil {
		t.Fatal(err)
	}
	defer m2.Shutdown()
	c2 := m2.Connected().First().AsChan()

	m3, err := New(cfg3)
	if err != nil {
		t.Fatal(err)
	}
	defer m3.Shutdown()
	c3 := m3.Connected().First().AsChan()

	m1.Run()
	m2.Run()
	m3.Run()

	select {
	case <-time.After(5 * time.Second):
		t.Error("peer2 did not connect", cfg2.UUID())
	case <-c2:
		if len(m2.peers) != 1 {
			t.Error("Not all peers connected to peer 2, want 1 got", len(m2.peers))
		}
	}

	select {
	case <-time.After(5 * time.Second):
		t.Error("peer3 did not connect", cfg3.UUID())
	case <-c3:
		if len(m3.peers) != 1 {
			t.Error("Not all peers connected to peer 3, want 1 got", len(m2.peers))
		}
	}

	time.Sleep(5 * time.Second)
	if len(m1.peers) != 2 {
		t.Error("Not all peers connected to peer 1, want 2 got", len(m1.peers))
	}
}

func TestManagerMessaging(t *testing.T) {
	cfg1 := config.New(os.Stdout, true)
	cfg1.OverrideUUID("exp")
	cfg1.AddOrSetUserTag("tag1", "1")
	cfg1.OverrideInterfaces([]string{"127.0.0.1"})

	cfg2 := config.New(os.Stdout, false)
	cfg2.OverrideUUID("imp1")
	cfg2.AddOrSetUserTag("tag1", "1")
	cfg2.OverrideInterfaces([]string{"127.0.0.1"})

	cfg3 := config.New(os.Stdout, false)
	cfg3.OverrideUUID("imp2")
	cfg3.AddOrSetUserTag("tag1", "1")
	cfg3.OverrideInterfaces([]string{"127.0.0.1"})

	m1, err := New(cfg1)
	if err != nil {
		t.Fatal(err)
	}
	defer m1.Shutdown()
	c1 := m1.Messages().Transform(connection.ToMessage).AsChan()

	m2, err := New(cfg2)
	if err != nil {
		t.Fatal(err)
	}
	defer m2.Shutdown()
	c2 := m2.Messages().Transform(connection.ToMessage).AsChan()

	m3, err := New(cfg3)
	if err != nil {
		t.Fatal(err)
	}
	defer m3.Shutdown()
	c3 := m3.Messages().Transform(connection.ToMessage).AsChan()

	m1.Run()
	f := m2.Connected().First()
	m2.Run()
	f.WaitUntilComplete()
	f = m3.Connected().First()
	m3.Run()
	f.WaitUntilComplete()
	msg := &messages.Mock{true}
	m2.Send(msg)

	if r := (<-c1).(*messages.Mock); r.Data != msg.Data {
		t.Error("peer1 did not not got the message", r)
	}
	time.Sleep(10 * time.Millisecond)
	m1.SendTo(cfg3.UUID(), msg)
	if r := (<-c3).(*messages.Mock); r.Data != msg.Data {
		t.Error("peer3 did not not got the message", r)
	}

	m1.SendToAll(msg)
	if r := (<-c2).(*messages.Mock); r.Data != msg.Data {
		t.Error("peer2 did not not got the message", r)
	}
	if r := (<-c3).(*messages.Mock); r.Data != msg.Data {
		t.Error("peer3 did not not got the message", r)
	}
}

func TestManagerSendGuaranteed(t *testing.T) {

	m1 := getTestManager(true)
	defer m1.Shutdown()
	m1.Run()

	msg := &messages.Mock{true}
	akn := m1.SendGuaranteed(msg)
	m2 := getTestManager(false)
	c2 := m2.Messages().AsChan()

	f := m2.Connected().First()
	m2.Run()
	f.WaitUntilComplete()

	select {
	case <-c2:
	case <-time.After(10 * time.Millisecond):
		t.Error("p2 didnt got msg")
	}

	m2.Shutdown()

	m3 := getTestManager(false)
	c3 := m3.Messages().Transform(connection.ToMessage).AsChan()
	m3.Run()

	if r := (<-c3).(*messages.Mock); r.Data != msg.Data {
		t.Error("peer1 did not not got the message", r)
	}
	akn.Complete(nil)
	m3.Shutdown()

	m4 := getTestManager(false)
	c4 := m4.Messages().AsChan()
	m4.Run()
	select {
	case <-c4:
		t.Error("p4 did got msg")
	case <-time.After(500 * time.Millisecond):
	}
	m4.Shutdown()
}

func getTestManager(e bool) (m *Manager) {
	cfg := config.New(os.Stdout, e)
	cfg.AddOrSetUserTag("tag1", "1")
	cfg.OverrideInterfaces([]string{"127.0.0.1"})

	m, _ = New(cfg)
	return
}
