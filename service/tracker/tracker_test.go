package tracker

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/joernweissenborn/thingiverse.io/service"
)

func TestJoin(t *testing.T) {

	cfg1 := config.New(os.Stdout, true)
	cfg1.AddOrSetUserTag("tag", "1")
	cfg2 := config.New(os.Stdout, false)
	cfg2.AddOrSetUserTag("tag", "2")

	p1 := 666
	p2 := 667

	t1, err := New("127.0.0.1", p1, cfg1)

	if err != nil {
		t.Fatal(err)
	}

	t2, err := New("127.0.0.1", p2, cfg2)

	if err != nil {
		t.Fatal(err)
	}

	c1 := t1.Join().AsChan()
	c2 := t2.Join().AsChan()

	t2.JoinCluster([]string{fmt.Sprintf("%s:%d", "127.0.0.1", t1.Port())})

	select {
	case <-time.After(10 * time.Second):
		t.Fatal("Couldnt find tracker 2")
	case n := <-c1:

		if n.UUID() != cfg2.UUID() {
			t.Error("Found wrong UUID",n.UUID().FullString(),cfg2.UUID().FullString())
		}

		if n.Node.Meta[0] != service.PROTOCOLL_SIGNATURE || !bytes.Equal(n.Node.Meta[1:3], port2byte(p2)) || n.Node.Meta[3] != 0 || string(n.Node.Meta[4:]) != "tag:2" {
			t.Error("Wrong Meta")
		}
	}

	select {
	case <-time.After(10 * time.Second):
		t.Fatal("Couldnt find tracker 1")
	case n := <-c2:
		if n.UUID() != cfg1.UUID() {
			t.Error("Found wrong UUID",n.UUID().FullString(),cfg2.UUID().FullString())
		}

		if n.Node.Meta[0] != service.PROTOCOLL_SIGNATURE || !bytes.Equal(n.Node.Meta[1:3], port2byte(p1)) || n.Node.Meta[3] != 1 || string(n.Node.Meta[4:]) != "tag:1" {
			t.Error("Wrong Meta", n.Node.Meta)
		}
	}
}

func TestAutoJoin(t *testing.T) {

	cfg1 := config.New(os.Stdout, true)
	cfg1.AddOrSetUserTag("tag", "1")
	cfg2 := config.New(os.Stdout, false)
	cfg2.AddOrSetUserTag("tag", "2")

	t1, err := New("127.0.0.1", 0, cfg1)
	if err != nil {
		t.Fatal(err)
	}

	t2, err := New("127.0.0.1", 0, cfg2)
	if err != nil {
		t.Fatal(err)
	}

	c1 := t1.Join().AsChan()
	c2 := t2.Join().AsChan()

	err = t1.StartAutoJoin()
	if err != nil {
		t.Fatal(err)
	}

	err = t2.StartAutoJoin()
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-time.After(10 * time.Second):
		t.Fatal("Couldnt find tracker 2")
	case <-c1:
	}

	select {
	case <-time.After(10 * time.Second):
		t.Fatal("Couldnt find tracker 1")
	case <-c2:
	}
}
func TestLeaveAndReconnect(t *testing.T) {

	cfg1 := config.New(os.Stdout, true)
	cfg1.AddOrSetUserTag("tag", "1")
	cfg2 := config.New(os.Stdout, false)
	cfg2.AddOrSetUserTag("tag", "2")

	t1, err := New("127.0.0.1", 0, cfg1)
	if err != nil {
		t.Fatal(err)
	}

	t2, err := New("127.0.0.1", 0, cfg2)
	if err != nil {
		t.Fatal(err)
	}

	c1 := t1.Join().AsChan()
	c2 := t1.Leave().First().AsChan()

	err = t1.StartAutoJoin()
	if err != nil {
		t.Fatal(err)
	}

	err = t2.StartAutoJoin()
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-time.After(5 * time.Second):
		t.Fatal("Service didnt join")
	case <-c1:
	}

	t2.Shutdown(nil)

	select {
	case <-time.After(5 * time.Second):
		t.Fatal("Service didnt leave")
	case <-c2:
	}

	cfg3 := config.New(os.Stdout, false)
	cfg3.AddOrSetUserTag("tag", "2")

	t3, err := New("127.0.0.1", 0, cfg3)
	if err != nil {
		t.Fatal(err)
	}

	err = t3.StartAutoJoin()
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-time.After(10 * time.Second):
		t.Fatal("Service didnt join")
	case <-c1:
	}

}
