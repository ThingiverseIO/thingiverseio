package thingiverseio_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/joernweissenborn/thingiverseio"
	"github.com/joernweissenborn/thingiverseio/config"
	"github.com/joernweissenborn/thingiverseio/service/messages"
)

func TestCall(t *testing.T) {
	i, err := thingiverseio.NewInputFromConfig(testConfig(false))
	if err != nil {
		t.Fatal(err)
	}
	defer i.Remove()
	e, err := thingiverseio.NewOutputFromConfig(testConfig(true))
	if err != nil {
		t.Fatal(err)
	}
	defer e.Remove()
	c := e.Requests().AsChan()

	f1 := i.Connected()
	f2 := e.Connected()
	e.Run()
	i.Run()
	f1.WaitUntilComplete()
	f2.WaitUntilComplete()
	params := []byte{4, 5, 63, 4}
	f := i.Call("SayHello", params)

	select {
	case <-time.After(5 * time.Second):
		t.Fatal("Didnt Got Request")
	case r := <-c:
		if r.Input != i.UUID() {
			t.Error("Wrong Import UUID", r.Input, i.UUID())
		}
		var res []byte
		r.Decode(&res)
		if !bytes.Equal(res, params) {
			t.Error("Wrong Params", r.Parameter(), params)
		}
		e.Reply(r, params)
	}

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("Didnt Got Request")
	case r := <-f.AsChan():
		if r.Output != e.UUID() {
			t.Error("Wrong Export UUID", r.Output, e.UUID())
		}

		var res []byte
		r.Decode(&res)
		if !bytes.Equal(res, params) {
			t.Error("Wrong Params", r.Parameter(), params)
		}
	}
}

func TestTrigger(t *testing.T) {
	i1, err := thingiverseio.NewInputFromConfig(testConfig(false))
	if err != nil {
		t.Fatal(err)
	}
	defer i1.Remove()
	i2, err := thingiverseio.NewInputFromConfig(testConfig(false))
	if err != nil {
		t.Fatal(err)
	}
	defer i2.Remove()
	e, err := thingiverseio.NewOutputFromConfig(testConfig(true))
	if err != nil {
		t.Fatal(err)
	}
	defer e.Remove()
	c := e.Requests().AsChan()

	c1 := i1.ListenResults().AsChan()
	c2 := i2.ListenResults().AsChan()

	f1 := i1.Connected()
	f2 := i2.Connected()
	f3 := e.Connected()

	e.Run()
	i2.Run()
	i1.Run()

	f1.WaitUntilComplete()
	f2.WaitUntilComplete()
	f3.WaitUntilComplete()

	i1.Listen("SayHello")
	i2.Listen("SayHello")
	time.Sleep(1 * time.Second)

	params := []byte{4, 5, 63, 4}
	i1.Trigger("SayHello", params)

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Didnt Got Request")
	case r := <-c:
		if r.Input != i1.UUID() {
			t.Error("Wrong Import UUID", r.Input, i1.UUID())
		}

		var res []byte
		r.Decode(&res)
		if !bytes.Equal(res, params) {
			t.Error("Wrong Params", r.Parameter(), params)
		}
		e.Reply(r, params)
	}

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Didnt Got Result 1")
	case r := <-c1:
		if r.Output != e.UUID() {
			t.Error("Wrong Export UUID", r.Output, e.UUID())
		}

		var res []byte
		r.Decode(&res)
		if !bytes.Equal(res, params) {
			t.Error("Wrong Params", r.Parameter(), params)
		}
	}
	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Didnt Got Result 2")
	case r := <-c2:
		if r.Output != e.UUID() {
			t.Error("Wrong Export UUID", r.Output, e.UUID())
		}

		var res []byte
		r.Decode(&res)
		if !bytes.Equal(res, params) {
			t.Error("Wrong Params", r.Parameter(), params)
		}
	}

}

func TestCallAll(t *testing.T) {
	i, err := thingiverseio.NewInputFromConfig(testConfig(false))
	if err != nil {
		t.Fatal(err)
	}
	defer i.Remove()

	e1, err := thingiverseio.NewOutputFromConfig(testConfig(true))
	if err != nil {
		t.Fatal(err)
	}
	defer e1.Remove()
	c1 := e1.Requests().AsChan()

	e2, err := thingiverseio.NewOutputFromConfig(testConfig(true))
	if err != nil {
		t.Fatal(err)
	}
	defer e2.Remove()
	c2 := e2.Requests().AsChan()

	f1 := i.Connected()
	f2 := e1.Connected()
	f3 := e2.Connected()

	i.Run()
	e1.Run()
	e2.Run()

	f1.WaitUntilComplete()
	f2.WaitUntilComplete()
	f3.WaitUntilComplete()

	params := []byte{4, 5, 63, 4}
	params1 := []byte{3}
	params2 := []byte{6}
	s := messages.NewResultStreamController()
	s1, s2 := s.Stream().Split(func(d *messages.Result) bool { return d.Output == e1.UUID() })
	rc1 := s1.AsChan()
	rc2 := s2.AsChan()
	i.CallAll("SayHello", params, s)
	select {
	case <-time.After(5 * time.Second):
		t.Fatal("Didnt Got Request 1")
	case r := <-c1:
		if r.Input != i.UUID() {
			t.Error("Wrong Import UUID 1", r.Input, i.UUID())
		}

		var res []byte
		r.Decode(&res)
		if !bytes.Equal(res, params) {
			t.Error("Wrong Params 1", r.Parameter(), params)
		}
		e1.Reply(r, params1)
	}

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("Didnt Got Request 2")
	case r := <-c2:
		if r.Input != i.UUID() {
			t.Error("Wrong Import UUID 2", r.Input, i.UUID())
		}

		var res []byte
		r.Decode(&res)
		if !bytes.Equal(res, params) {
			t.Error("Wrong Params 2", r.Parameter(), params)
		}
		e2.Reply(r, params2)
	}

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("Didnt Got Result 1")
	case r := <-rc1:
		if r.Output != e1.UUID() {
			t.Error("Wrong Export UUID", r.Output, e1.UUID())
		}

		var res []byte
		r.Decode(&res)
		if !bytes.Equal(res, params1) {
			t.Error("Wrong Params", r.Parameter(), params1)
		}
	}
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("Didnt Got Result 2")
	case r := <-rc2:
		if r.Output != e2.UUID() {
			t.Error("Wrong Export UUID", r.Output, e2.UUID())
		}

		var res []byte
		r.Decode(&res)
		if !bytes.Equal(res, params2) {
			t.Error("Wrong Params", r.Parameter(), params2)
		}
	}

}

func TestEmit(t *testing.T) {
	i, err := thingiverseio.NewInputFromConfig(testConfig(false))
	if err != nil {
		t.Fatal(err)
	}
	e, err := thingiverseio.NewOutputFromConfig(testConfig(true))
	if err != nil {
		t.Fatal(err)
	}
	defer e.Remove()
	c := i.ListenResults().AsChan()
	i.Listen("SayHello")
	f1 := i.Connected()
	f2 := e.Connected()

	i.Run()
	e.Run()
	f1.WaitUntilComplete()
	f2.WaitUntilComplete()
	time.Sleep(500 * time.Millisecond)

	params := []byte{4, 5, 63, 4}
	e.Emit("SayHello", nil, params)

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("Didnt Got Result")
	case r := <-c:
		if r.Output != e.UUID() {
			t.Error("Wrong Export UUID", r.Output, e.UUID())
		}

		var res []byte
		r.Decode(&res)
		if !bytes.Equal(res, params) {
			t.Error("Wrong Params", r.Parameter(), params)
		}
	}
	i.StopListen("SayHello")
	time.Sleep(100 * time.Millisecond)
	e.Emit("SayHello", nil, params)

	select {
	case <-time.After(2 * time.Microsecond):
	case <-c:
		t.Fatal("Got Result")
	}
	i.Remove()
	time.Sleep(1 * time.Second)
	//testing if it not crashes due to uncomplete removal
	e.Emit("SayHello", nil, params)
}

var testDescriptor = &thingiverseio.Descriptor{
	[]thingiverseio.Function{
		thingiverseio.Function{
			Name: "SayHello",
			Input: []thingiverseio.Parameter{
				thingiverseio.Parameter{
					Name: "Greeting",
					Type: "string"}},
			Output: []thingiverseio.Parameter{
				thingiverseio.Parameter{
					Name: "Answer",
					Type: "string"}},
		},
	},
	map[string]string{"TAG_1": ""},
}

func testConfig(export bool) (cfg *config.Config) {
	cfg = config.New(export, testDescriptor.AsTagSet())
	return
}
