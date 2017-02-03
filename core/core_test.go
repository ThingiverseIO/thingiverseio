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

func TestCallGuarantee(t *testing.T) {

	desc, _ := descriptor.Parse(Descriptor1)

	cfg := config.DefaultLocalhost()

	mt1 := &network.MockTracker{}
	mt2 := &network.MockTracker{}
	mt3 := &network.MockTracker{}
	mps := network.NewMockProvider(3)

	i, _ := core.NewInputCore(desc, cfg, mt1, mps[0])
	o, _ := core.NewOutputCore(desc, cfg, mt2, mps[1])

	arr := network.Arrival{
		IsOutput: true,
		Details:  mt2.Dt,
		UUID:     o.UUID(),
	}

	mt1.Av.Add(arr)
	i.ConnectedFuture().WaitUntilTimeout(100 * time.Millisecond)
	o.ConnectedFuture().WaitUntilTimeout(100 * time.Millisecond)

	data := []byte{1, 2, 3, 4}
	result, uuid, _ := i.Request("testfun", message.CALL, data)

	if len(i.Pending()) != 1 {
		t.Error("Request did not got registered as pending.")
	}

	time.Sleep(1 * time.Millisecond)

	if i.Pending()[uuid].Output != o.UUID() {
		t.Error("Request did not got registered with right output id:", i.Pending()[uuid].Output, o.UUID())
	}

	o.Shutdown()

	time.Sleep(1 * time.Millisecond)

	if !i.Pending()[uuid].Output.IsEmpty() {
		t.Error("Output did not got deregistered from pending")
	}

	o2, _ := core.NewOutputCore(desc, cfg, mt3, mps[2])

	request := o2.RequestStream().First()
	arr = network.Arrival{
		IsOutput: true,
		Details:  mt3.Dt,
		UUID:     o2.UUID(),
	}

	mt1.Av.Add(arr)

	if !request.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Request did not arrive")
	}

	if i.Pending()[uuid].Output != o2.UUID() {
		t.Error("Request did not got reregistered with right output id:", i.Pending()[uuid].Output, o.UUID())
	}
	data2 := []byte{4, 5, 6, 7}
	o.Reply(request.Result(), data2)

	if !result.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Result did not arrive")
	}

	time.Sleep(1 * time.Millisecond)

	if len(i.Pending()) != 0 {
		t.Error("Request did not got deregistered from pending.")
	}

}

func TestTrigger(t *testing.T) {

	desc, _ := descriptor.Parse(Descriptor1)

	i, o := getInputOutput(desc, desc)

	request := o.RequestStream().First()
	result := i.ListenStream().First()
	data := []byte{1, 2, 3, 4}
	_, uuid, err := i.Request("testfun", message.TRIGGER, data)
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

	if result.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Result did arrive")
	}

	i.StartListen("testfun")
	time.Sleep(1 * time.Millisecond)

	request = o.RequestStream().First()
	_, uuid, _ = i.Request("testfun", message.TRIGGER, data)

	if !request.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Request did not arrive")
	}

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

	i.StopListen("testfun")
	time.Sleep(1 * time.Millisecond)

	request = o.RequestStream().First()
	result = i.ListenStream().First()
	_, uuid, _ = i.Request("testfun", message.TRIGGER, data)

	if !request.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Request did not arrive")
	}

	o.Reply(request.Result(), data2)

	if result.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Result did arrive")
	}
}

func TestTriggerAll(t *testing.T) {

	desc, _ := descriptor.Parse(Descriptor1)

	cfg := config.DefaultLocalhost()

	mt1 := &network.MockTracker{}
	mt2 := &network.MockTracker{}
	mt3 := &network.MockTracker{}
	mps := network.NewMockProvider(3)

	i, _ := core.NewInputCore(desc, cfg, mt1, mps[0])
	o1, _ := core.NewOutputCore(desc, cfg, mt2, mps[1])
	o2, _ := core.NewOutputCore(desc, cfg, mt3, mps[2])

	arr := network.Arrival{
		IsOutput: true,
		Details:  mt2.Dt,
		UUID:     o1.UUID(),
	}

	mt1.Av.Add(arr)

	arr = network.Arrival{
		IsOutput: true,
		Details:  mt3.Dt,
		UUID:     o2.UUID(),
	}

	mt1.Av.Add(arr)

	if !i.ConnectedFuture().WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Input did not connect.")
	}
	o1.ConnectedFuture().WaitUntilTimeout(100 * time.Millisecond)
	o2.ConnectedFuture().WaitUntilTimeout(100 * time.Millisecond)

	i.StartListen("testfun")

	request1 := o1.RequestStream().First()
	request2 := o2.RequestStream().First()
	result := i.ListenStream().First()
	data := []byte{1, 2, 3, 4}
	_, _, err := i.Request("testfun", message.TRIGGERALL, data)
	if err != nil {
		t.Fatal(err)
	}

	if !request1.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Request1 did not arrive")
	}

	data2 := []byte{4, 5, 6, 7}
	o1.Reply(request1.Result(), data2)

	if !result.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Result 1 did not arrive")
	}

	result = i.ListenStream().First()

	if !request2.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Request1 did not arrive")
	}

	o2.Reply(request2.Result(), data2)

	if !result.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Result 2 did not arrive")
	}

}
