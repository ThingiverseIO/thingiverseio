package thingiverseio_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/ThingiverseIO/thingiverseio"
	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/message"
)

var testDescriptor = `
function SayHello(Greeting string) (Answer string)
property Mood: Mood string
tag TAG_1
`

func testConfig() (cfg *config.UserConfig) {
	cfg = config.DefaultLocalhost()
	return
}

func TestObserveProperty(t *testing.T) {
	i, err := thingiverseio.NewInputFromConfig(testDescriptor, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer i.Remove()
	o, err := thingiverseio.NewOutputFromConfig(testDescriptor, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer o.Remove()

	testprop := []byte{1, 9, 5}

	if err := i.StartObservation("Mood"); err != nil {
		t.Fatal(err)
	}

	f1 := i.ConnectedFuture()
	f2 := o.ConnectedFuture()
	o.Run()
	i.Run()
	f1.WaitUntilComplete()
	f2.WaitUntilComplete()
	if err := o.SetProperty("Mood", testprop); err != nil {
		t.Fatal("Failed to set property", err)
	}

	time.Sleep(100 * time.Millisecond)
	p, err := i.GetProperty("Mood")
	if err != nil {
		t.Fatal(err)
	}
	var v []byte
	err = p.Value(&v)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(v, testprop) {
		t.Error("wrong property value", v, testprop)
	}
}

func TestUpdateProperty(t *testing.T) {
	i, err := thingiverseio.NewInputFromConfig(testDescriptor, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer i.Remove()
	o, err := thingiverseio.NewOutputFromConfig(testDescriptor, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer o.Remove()

	testprop := []byte{4, 5, 63, 4}
	if err := o.SetProperty("Mood", testprop); err != nil {
		t.Fatal(err)
	}

	f1 := i.ConnectedFuture()
	f2 := o.ConnectedFuture()
	o.Run()
	i.Run()
	f1.WaitUntilComplete()
	f2.WaitUntilComplete()

	f, err := i.UpdateProperty("Mood")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-time.After(100 * time.Millisecond):
		t.Error("didn't received property")
	case p := <-f.AsChan():
		if p.Name != "Mood" {
			t.Error("Wrong name: Mood", p.Name)
		}
		var v []byte
		if err = p.Value(&v); err != nil {
			t.Error(err)
		}
		if !bytes.Equal(v, testprop) {
			t.Error("wrong property value", v, testprop)
		}
	}
}

func TestCall(t *testing.T) {
	i, err := thingiverseio.NewInputFromConfig(testDescriptor, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer i.Remove()
	e, err := thingiverseio.NewOutputFromConfig(testDescriptor, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer e.Remove()
	c, _ := e.Requests().AsChan()

	f1 := i.ConnectedFuture()
	f2 := e.ConnectedFuture()
	e.Run()
	i.Run()
	f1.WaitUntilComplete()
	f2.WaitUntilComplete()
	params := []byte{4, 5, 63, 4}
	f, _ := i.Call("SayHello", params)

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
	i1, err := thingiverseio.NewInputFromConfig(testDescriptor, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer i1.Remove()
	i2, err := thingiverseio.NewInputFromConfig(testDescriptor, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer i2.Remove()
	e, err := thingiverseio.NewOutputFromConfig(testDescriptor, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer e.Remove()
	c, _ := e.Requests().AsChan()

	c1, _ := i1.ListenResults().AsChan()
	c2, _ := i2.ListenResults().AsChan()

	f1 := i1.ConnectedFuture()
	f2 := i2.ConnectedFuture()
	f3 := e.ConnectedFuture()

	e.Run()
	i2.Run()
	i1.Run()

	i1.StartListen("SayHello")
	i2.StartListen("SayHello")

	f1.WaitUntilComplete()
	f2.WaitUntilComplete()
	f3.WaitUntilComplete()

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
	i, err := thingiverseio.NewInputFromConfig(testDescriptor, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer i.Remove()

	e1, err := thingiverseio.NewOutputFromConfig(testDescriptor, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer e1.Remove()
	c1, _ := e1.Requests().AsChan()

	e2, err := thingiverseio.NewOutputFromConfig(testDescriptor, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer e2.Remove()
	c2, _ := e2.Requests().AsChan()

	f1 := i.ConnectedFuture()
	f2 := e1.ConnectedFuture()
	f3 := e2.ConnectedFuture()

	i.Run()
	e1.Run()
	e2.Run()

	f1.WaitUntilComplete()
	f2.WaitUntilComplete()
	f3.WaitUntilComplete()

	time.Sleep(1 * time.Second)

	params := []byte{4, 5, 63, 4}
	params1 := []byte{3}
	params2 := []byte{6}
	s, _ := i.CallAll("SayHello", params)
	s1, s2 := s.Split(func(d *message.Result) bool { return d.Output == e1.UUID() })
	rc1, _ := s1.AsChan()
	rc2, _ := s2.AsChan()
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
	i, err := thingiverseio.NewInputFromConfig(testDescriptor, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	e, err := thingiverseio.NewOutputFromConfig(testDescriptor, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer e.Remove()
	c, _ := i.ListenResults().AsChan()
	i.StartListen("SayHello")
	f1 := i.ConnectedFuture()
	f2 := e.ConnectedFuture()

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
