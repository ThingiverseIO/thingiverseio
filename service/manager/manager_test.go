package manager

import (
	"os"
	"testing"
	"time"

	"github.com/joernweissenborn/thingiverse.io/config"
)

func TestManager(t *testing.T) {

	cfg1 := config.New(os.Stdout, true)
	cfg1.AddOrSetUserTag("tag1", "1")
	cfg1.AddOrSetUserTag("tag2", "2")
	cfg1.OverrideInterfaces([]string{"127.0.0.1"})

	cfg2 := config.New(os.Stdout, false)
	cfg2.AddOrSetUserTag("tag2", "2")
	cfg2.OverrideInterfaces([]string{"127.0.0.1"})

	cfg3 := config.New(os.Stdout, false)
	cfg3.AddOrSetUserTag("tag1", "1")
	cfg3.OverrideInterfaces([]string{"127.0.0.1"})

	m1, err := New(cfg1)
	if err != nil {
		t.Fatal(err)
	}
	c1 := m1.Connected().First().AsChan()

	m2, err := New(cfg2)
	if err != nil {
		t.Fatal(err)
	}
	c2 := m2.Connected().First().AsChan()

	m3, err := New(cfg3)
	if err != nil {
		t.Fatal(err)
	}
	c3 := m3.Connected().First().AsChan()

	m1.Run()
	m2.Run()
	m3.Run()

	select {
	case <-time.After(10 * time.Second):
		t.Fatal("peer2 did not connect",cfg2.UUID())
	case <-c2:
		if len(m2.peers) != 1 {
			t.Error("Not all peers connected to peer 2, want 1 got", len(m2.peers))
		}
	}

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("peer3 did not connect")
	case <-c3:
		if len(m3.peers) != 1 {
			t.Error("Not all peers connected to peer 3, want 1 got", len(m2.peers))
		}
	}

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("peer1 did not connect")
	case <-c1:
		if len(m1.peers) != 2 {
			t.Error("Mot all peers connected to peer 1, want 2 got", len(m1.peers))
		}
	}
}
