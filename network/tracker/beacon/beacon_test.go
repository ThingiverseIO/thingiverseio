package beacon

import (
	"testing"
	"time"
)

func testconf(payload []byte) *Config {
	return &Config{
		Address:      "127.0.0.1",
		Port:         6660,
		PingInterval: 1 * time.Millisecond,
		Payload:      payload,
	}
}

func TestBeacon(t *testing.T) {
	p1 := []byte("1234")
	b1, err := New(testconf(p1))
	if err != nil {
		t.Fatal(err)
	}
	defer b1.Stop()
	c1 := b1.Signals().First().AsChan()

	p2 := []byte("HALLO")
	b2, err := New(testconf(p2))
	if err != nil {
		t.Fatal(err)
	}
	defer b2.Stop()
	c2 := b2.Signals().First().AsChan()
	b1.Run()
	b2.Run()
	b1.Ping()
	b2.Ping()
	select {
	case data := <-c1:
		if string(data.Data) != "HALLO" {
			t.Error("Wrong data, needed 'HALLO', got", data)
		}
	case <-time.After(1000 * time.Millisecond):
		t.Error("Didnt got network.beacon 2")
	}

	select {
	case data := <-c2:
		if string(data.Data) != "1234" {
			t.Error("Wrong data, needed '1234', got", data)
		}
	case <-time.After(1000 * time.Millisecond):
		t.Error("Didn't got network.beacon 1")
	}
}

func TestBeaconstop(t *testing.T) {
	p1 := []byte("1234")
	b1, err := New(testconf(p1))
	if err != nil {
		t.Fatal(err)
	}
	defer b1.Stop()
	p2 := []byte("HALLO")
	b2, err := New(testconf(p2))
	if err != nil {
		t.Fatal(err)
	}
	b2.Run()
	b1.Run()
	b1.Ping()
	b2.Ping()
	time.Sleep(10 * time.Millisecond)
	b2.Stop()
	time.Sleep(10 * time.Millisecond)
	c1 := b1.Signals().First().AsChan()
	time.Sleep(10 * time.Millisecond)
	select {
	case <-c1:
		t.Error("Becaon didnt stop")
	case <-time.After(1 * time.Microsecond):

	}
}
