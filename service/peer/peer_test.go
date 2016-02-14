package peer

import (
	"os"
	"testing"
	"time"

	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/joernweissenborn/thingiverse.io/service/connection"
	"github.com/joernweissenborn/thingiverse.io/service/messages"
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

	cfg1 := config.New(os.Stdout, true)
	cfg2 := config.New(os.Stdout, false)

	c := i2.In().Where(connection.IsMsgFromSender(cfg1.UUID())).Where(validMsg).Transform(transformToMessage).Where(messages.Is(messages.HELLO)).AsChan()

	p1, err := New(cfg2.UUID(), "127.0.0.1", i2.Port(), i1, cfg1)
	if err != nil {
		t.Fatal(err)
	}
	p1.InitConnection()

	var p2 *Peer
	select {
	case <-time.After(10 * time.Second):
		t.Fatal("Did not received Hello")
	case d := <-c:
		m := d.(*messages.Hello)
		if m.Address != i1.Addr() || m.Port != i1.Port() {
			t.Fatal("Wrong message", m)
		}
		p2, err = NewFromHello(cfg1.UUID(), m, i2, cfg2)
		if err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(10 * time.Second)

	if !p1.initialized.Completed() || !p2.initialized.Completed() {
		t.Error("Connection did not initialize")
	}

}
