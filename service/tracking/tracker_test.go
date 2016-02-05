package tracking

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/joernweissenborn/thingiverse.io/config"
)

func TestJoinAndLeave(t *testing.T) {

	cfg1 := config.New(os.Stdout, true)
	cfg1.AddUserTag("tag", "1")
	cfg2 := config.New(os.Stdout, false)
	cfg2.AddUserTag("tag", "2")

	t1, err := Create("127.0.0.1", cfg1)

	if err != nil {
		t.Fatal(err)
	}

	t2, err := Create("127.0.0.1", cfg2)

	if err != nil {
		t.Fatal(err)
	}

	c1 := t1.Join().AsChan()
	c2 := t2.Join().AsChan()

	t2.joinCluster([]string{fmt.Sprintf("%s:%d", "127.0.0.1", t1.Port())})

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Couldnt find tracker 2")
	case data := <-c1:
		n := data.(*memberlist.Node)

		if !strings.Contains(n.Name, cfg2.UUID()) {
			t.Error("Found wrong UUID")
		}

		if n.Meta[0] != 0xA5 || n.Meta[1] != 0 || string(n.Meta[2:]) != "tag2" {
			t.Error("Wrong Meta", n.Meta)
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

		if n.Meta[0] != 0xA5 || n.Meta[1] != 1 || string(n.Meta[2:]) != "tag1" {
			t.Error("Wrong Meta", n.Meta)
		}
	}
}
