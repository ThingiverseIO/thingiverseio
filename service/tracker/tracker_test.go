package tracker

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/joernweissenborn/thingiverse.io/service"
)

func TestJoin(t *testing.T) {

	cfg1 := config.New(os.Stdout, true)
	cfg1.AddUserTag("tag", "1")
	cfg2 := config.New(os.Stdout, false)
	cfg2.AddUserTag("tag", "2")

	p1 := 666
	p2 := 667

	t1, err := Create("127.0.0.1", p1, cfg1)

	if err != nil {
		t.Fatal(err)
	}

	t2, err := Create("127.0.0.1", p2, cfg2)

	if err != nil {
		t.Fatal(err)
	}

	c1 := t1.Join().AsChan()
	c2 := t2.Join().AsChan()

	t2.JoinCluster([]string{fmt.Sprintf("%s:%d", "127.0.0.1", t1.Port())})

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Couldnt find tracker 2")
	case data := <-c1:
		n := data.(*memberlist.Node)

		if !strings.Contains(n.Name, cfg2.UUID()) {
			t.Error("Found wrong UUID")
		}

		if n.Meta[0] != service.PROTOCOLL_SIGNATURE || !bytes.Equal(n.Meta[1:3], port2byte(p2)) || n.Meta[3] != 0 || string(n.Meta[4:]) != "tag2" {
			t.Error("Wrong Meta")
		}
	}

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Couldnt find tracker 1")
	case data := <-c2:
		n := data.(*memberlist.Node)
		if !strings.Contains(n.Name, cfg1.UUID()) {
			t.Error("Found wrong UUID")
		}

		if n.Meta[0] != service.PROTOCOLL_SIGNATURE || !bytes.Equal(n.Meta[1:3], port2byte(p1)) || n.Meta[3] != 1 || string(n.Meta[4:]) != "tag1" {
			t.Error("Wrong Meta", n.Meta)
		}
	}
}

func TestAutoJoin(t *testing.T) {

	cfg1 := config.New(os.Stdout, true)
	cfg1.AddUserTag("tag", "1")
	cfg2 := config.New(os.Stdout, false)
	cfg2.AddUserTag("tag", "2")

	t1, err := Create("127.0.0.1", 0, cfg1)
	if err != nil {
		t.Fatal(err)
	}

	t2, err := Create("127.0.0.1", 0, cfg2)
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
	case <-time.After(5 * time.Second):
		t.Fatal("Couldnt find tracker 2")
	case <-c1:
	}

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Couldnt find tracker 1")
	case <-c2:
	}
}
func TestLeaveAndReconnect(t *testing.T) {

	cfg1 := config.New(os.Stdout, true)
	cfg1.AddUserTag("tag", "1")
	cfg2 := config.New(os.Stdout, false)
	cfg2.AddUserTag("tag", "2")

	t1, err := Create("127.0.0.1", 0, cfg1)
	if err != nil {
		t.Fatal(err)
	}

	t2, err := Create("127.0.0.1", 0, cfg2)
	if err != nil {
		t.Fatal(err)
	}

	c1 := t1.Join().AsChan()
	c2 := t2.Leave().First().AsChan()

	err = t1.StartAutoJoin()
	if err != nil {
		t.Fatal(err)
	}

	err = t2.StartAutoJoin()
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Service didnt join")
	case <-c1:
	}

	t2.Stop()

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Service didnt join")
	case data := <-c2:
		n := data.(*memberlist.Node)
		if !strings.Contains(n.Name, cfg2.UUID()) {
			t.Error("Found wrong UUID", n.Name)
		}
	}

	cfg3 := config.New(os.Stdout, false)
	cfg3.AddUserTag("tag", "2")

	t3, err := Create("127.0.0.1", 0, cfg3)
	if err != nil {
		t.Fatal(err)
	}

	err = t3.StartAutoJoin()
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Service didnt join")
	case <-c1:
	}

}
