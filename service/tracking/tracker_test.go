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

	cfg1 := config.New(os.Stdout)
	cfg2 := config.New(os.Stdout)

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
		if !strings.Contains(data.(*memberlist.Node).Name, cfg2.UUID()) {
			t.Error("Found wrong UUID")
		}
	}

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Couldnt find tracker 1")
	case data := <-c2:
		if !strings.Contains(data.(*memberlist.Node).Name, cfg1.UUID()) {
			t.Error("Found wrong UUID")
		}
	}
}
