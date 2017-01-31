package core_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/core"
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ThingiverseIO/thingiverseio/message"
	"github.com/ThingiverseIO/thingiverseio/network"
)

var Descriptor1 = `
func testfun()()
tags TAG1, TEST:2
`

var Descriptor2 = "tag TEST:2"

func getInputOutput(descIn, descOut descriptor.Descriptor) (i core.InputCore, o core.OutputCore) {

	cfg := config.DefaultLocalhost()

	mt1 := &network.MockTracker{}
	mt2 := &network.MockTracker{}
	mps := network.NewMockProvider(2)

	i, _ = core.NewInputCore(descIn, cfg, mt1, mps[0])
	o, _ = core.NewOutputCore(descOut, cfg, mt2, mps[1])

	arr := network.Arrival{
		IsOutput: true,
		Details:  mt2.Dt,
		UUID:     o.UUID(),
	}

	mt1.Av.Add(arr)
	i.ConnectedFuture().WaitUntilTimeout(100 * time.Millisecond)
	o.ConnectedFuture().WaitUntilTimeout(100 * time.Millisecond)

	return
}

func TestBasicConnection(t *testing.T) {

	desc, _ := descriptor.Parse(Descriptor1)
	cfg := config.DefaultLocalhost()

	mt1 := &network.MockTracker{}
	mt2 := &network.MockTracker{}
	mps := network.NewMockProvider(2)

	i, _ := core.NewInputCore(desc, cfg, mt1, mps[0])
	o, _ := core.NewOutputCore(desc, cfg, mt2, mps[1])

	arr := network.Arrival{
		IsOutput: true,
		Details:  mt2.Dt,
		UUID:     o.UUID(),
	}

	mt1.Av.Add(arr)

	if !i.ConnectedFuture().WaitUntilTimeout(100*time.Millisecond) ||
		!o.ConnectedFuture().WaitUntilTimeout(100*time.Millisecond) {
		t.Fatal("Peers didn't connect.")
	}

	mt1.Lv.Add(o.UUID())
	mt2.Lv.Add(i.UUID())

	time.Sleep(10 * time.Millisecond)

	if i.Connected() || o.Connected() {
		t.Fatal("Peers didn't disconnect.", i.Connected())
	}
}

func TestNonMatchingDescriptor(t *testing.T) {

	desc1, _ := descriptor.Parse(Descriptor1)
	desc2, _ := descriptor.Parse(Descriptor2)
	i, o := getInputOutput(desc1, desc2)

	if i.Connected() || o.Connected() {
		t.Fatal("Peers did connect.")
	}
}

func TestCall(t *testing.T) {

	desc, _ := descriptor.Parse(Descriptor1)

	i, o := getInputOutput(desc, desc)

	request := o.RequestStream().First()
	data := []byte{1, 2, 3, 4}
	result, uuid, err := i.Request("testfun", message.CALL, data)
	if err != nil {
		t.Fatal(err)
	}

	if !request.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Request did not arrive")
	}

	if request.Result().UUID != uuid {
		t.Fatal("Wrong Request UUID", request.Result().UUID, uuid)
	}

	if !bytes.Equal(data, request.Result().Parameter()) {
		t.Fatal("Wrong request parameter", data, request.Result().Parameter())
	}

	data2 := []byte{4, 5, 6, 7}
	o.Reply(request.Result(), data2)

	if !result.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Result did not arrive")
	}

	if result.Result().Request.UUID != uuid {
		t.Fatal("Wrong Request UUID", request.Result().UUID, uuid)
	}

	if !bytes.Equal(data2, result.Result().Parameter()) {
		t.Fatal("Wrong result parameter", data2, result.Result().Parameter())
	}

}
