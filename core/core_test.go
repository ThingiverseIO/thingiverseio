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
function testfun()()
property testprop: data bin
stream teststream: data bin
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
	i.ConnectedObservable().Stream().First().WaitUntilTimeout(500 * time.Millisecond)
	o.ConnectedObservable().Stream().First().WaitUntilTimeout(500 * time.Millisecond)

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

	defer i.Shutdown()
	defer o.Shutdown()

	arr := network.Arrival{
		IsOutput: true,
		Details:  mt2.Dt,
		UUID:     o.UUID(),
	}
	f1 := i.ConnectedObservable().Stream().First()
	f2 := o.ConnectedObservable().Stream().First()
	mt1.Av.Add(arr)

	if !f1.WaitUntilTimeout(100*time.Millisecond) ||
		!f2.WaitUntilTimeout(100*time.Millisecond) {
		t.Fatal("Peers didn't connect.")
	}

	f1 = i.ConnectedObservable().Stream().First()
	f2 = o.ConnectedObservable().Stream().First()
	mt1.Lv.Add(o.UUID())
	mt2.Lv.Add(i.UUID())

	//time.Sleep(10 * time.Millisecond)

	if !f1.WaitUntilTimeout(100*time.Millisecond) ||
		!f2.WaitUntilTimeout(100*time.Millisecond) {
		t.Fatal("Peers didn't disconnect.", i.Connected())
	}
}

func TestNonMatchingDescriptor(t *testing.T) {

	desc1, _ := descriptor.Parse(Descriptor1)
	desc2, _ := descriptor.Parse(Descriptor2)
	i, o := getInputOutput(desc1, desc2)
	defer i.Shutdown()
	defer o.Shutdown()

	if i.Connected() || o.Connected() {
		t.Fatal("Peers did connect.")
	}
}

func TestStreams(t *testing.T) {

	desc, _ := descriptor.Parse(Descriptor1)

	i, o := getInputOutput(desc, desc)
	defer i.Shutdown()
	defer o.Shutdown()

	testprop := []byte{1, 9, 5}

	i.StartConsume("teststream")

	s, err := i.GetStream("teststream")
	if err != nil {
		t.Fatal("Error getting stream", err)
	}
	f := s.First()
	time.Sleep(100 * time.Millisecond)

	if err := o.AddStream("teststream", testprop); err != nil {
		t.Fatal("Failed adding to stream", err)
	}

	time.Sleep(100 * time.Millisecond)
	if !f.Completed() {
		t.Fatal("Stream event didn't arrive")
	}

	if !bytes.Equal(testprop, f.Result().([]byte)) {
		t.Error("wrong stream event value", f.Result(), testprop)
	}

	f = s.First()
	if err = i.StopConsume("teststream");err!=nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)

	if err := o.AddStream("teststream", testprop); err != nil {
		t.Fatal("Failed adding to stream", err)
	}

	time.Sleep(100 * time.Millisecond)
	if f.Completed() {
		t.Fatal("Stream event did arrive")
	}
}

func TestObserveProperty(t *testing.T) {

	desc, _ := descriptor.Parse(Descriptor1)

	i, o := getInputOutput(desc, desc)

	testprop := []byte{1, 9, 5}

	i.StartObservation("testprop")

	time.Sleep(100 * time.Millisecond)

	if err := o.SetProperty("testprop", testprop); err != nil {
		t.Fatal("Failed to set property", err)
	}

	time.Sleep(100 * time.Millisecond)
	v, err := i.GetProperty("testprop")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(testprop, v.Value().([]byte)) {
		t.Error("wrong property value", v, testprop)
	}
	defer i.Shutdown()
	defer o.Shutdown()
}

func TestUpdateProperty(t *testing.T) {

	desc, _ := descriptor.Parse(Descriptor1)

	i, o := getInputOutput(desc, desc)

	testprop := []byte{1, 9, 5}

	if err := o.SetProperty("testprop", testprop); err != nil {
		t.Fatal("Failed to set property", err)
	}

	time.Sleep(1 * time.Millisecond)
	f, err := i.UpdateProperty("testprop")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-time.After(100 * time.Millisecond):
		t.Error("didn't received property")
	case v := <-f.AsChan():
		if !bytes.Equal(v.([]byte), testprop) {
			t.Error("wrong property value", v, testprop)
		}
	}

	defer i.Shutdown()
	defer o.Shutdown()
}

func TestCall(t *testing.T) {

	desc, _ := descriptor.Parse(Descriptor1)

	i, o := getInputOutput(desc, desc)

	defer i.Shutdown()
	defer o.Shutdown()
	request := o.RequestStream().First()
	data := []byte{1, 2, 3, 4}
	result, _, uuid, err := i.Request("testfun", message.CALL, data)
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
	defer i.Shutdown()
	o, _ := core.NewOutputCore(desc, cfg, mt2, mps[1])
	arr := network.Arrival{
		IsOutput: true,
		Details:  mt2.Dt,
		UUID:     o.UUID(),
	}

	mt1.Av.Add(arr)
	i.ConnectedObservable().Stream().First().WaitUntilTimeout(100 * time.Millisecond)
	o.ConnectedObservable().Stream().First().WaitUntilTimeout(100 * time.Millisecond)

	data := []byte{1, 2, 3, 4}
	result, _, _, _ := i.Request("testfun", message.CALL, data)

	time.Sleep(10 * time.Millisecond)

	o.Shutdown()

	time.Sleep(100 * time.Millisecond)

	o2, _ := core.NewOutputCore(desc, cfg, mt3, mps[2])
	defer o2.Shutdown()

	request := o2.RequestStream().First()
	arr = network.Arrival{
		IsOutput: true,
		Details:  mt3.Dt,
		UUID:     o2.UUID(),
	}

	mt1.Av.Add(arr)
	o2.ConnectedObservable().Stream().First().WaitUntilTimeout(100 * time.Millisecond)

	if !request.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Request did not arrive")
	}

	data2 := []byte{4, 5, 6, 7}
	o2.Reply(request.Result(), data2)

	if !result.WaitUntilTimeout(1000 * time.Millisecond) {
		t.Fatal("Result did not arrive")
	}

}

func TestTrigger(t *testing.T) {

	desc, _ := descriptor.Parse(Descriptor1)

	i, o := getInputOutput(desc, desc)

	defer i.Shutdown()
	defer o.Shutdown()
	request := o.RequestStream().First()
	result := i.ListenStream().First()
	data := []byte{1, 2, 3, 4}
	_, _, uuid, err := i.Request("testfun", message.TRIGGER, data)
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
	time.Sleep(100 * time.Millisecond)

	request = o.RequestStream().First()
	_, _, uuid, _ = i.Request("testfun", message.TRIGGER, data)

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
	_, _, uuid, _ = i.Request("testfun", message.TRIGGER, data)

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

	defer i.Shutdown()
	defer o1.Shutdown()
	defer o2.Shutdown()

	i.StartListen("testfun")

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

	if !i.ConnectedObservable().Stream().First().WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Input did not connect.")
	}
	i.ConnectedObservable().Stream().First().WaitUntilTimeout(100 * time.Millisecond)
	o1.ConnectedObservable().Stream().First().WaitUntilTimeout(100 * time.Millisecond)
	o2.ConnectedObservable().Stream().First().WaitUntilTimeout(100 * time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	request1 := o1.RequestStream().First()
	request2 := o2.RequestStream().First()
	result := i.ListenStream().First()
	data := []byte{1, 2, 3, 4}
	_, _, _, err := i.Request("testfun", message.TRIGGERALL, data)
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

func TestCallAll(t *testing.T) {

	desc, _ := descriptor.Parse(Descriptor1)

	cfg := config.DefaultLocalhost()

	mt1 := &network.MockTracker{}
	mt2 := &network.MockTracker{}
	mt3 := &network.MockTracker{}
	mps := network.NewMockProvider(3)

	i, _ := core.NewInputCore(desc, cfg, mt1, mps[0])
	o1, _ := core.NewOutputCore(desc, cfg, mt2, mps[1])
	o2, _ := core.NewOutputCore(desc, cfg, mt3, mps[2])

	defer i.Shutdown()
	defer o1.Shutdown()
	defer o2.Shutdown()
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

	if !i.ConnectedObservable().Stream().First().WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Input did not connect.")
	}
	o1.ConnectedObservable().Stream().First().WaitUntilTimeout(100 * time.Millisecond)
	o2.ConnectedObservable().Stream().First().WaitUntilTimeout(100 * time.Millisecond)

	request1 := o1.RequestStream().First()
	request2 := o2.RequestStream().First()
	data := []byte{1, 2, 3, 4}
	_, resultStream, _, err := i.Request("testfun", message.CALLALL, data)
	if err != nil {
		t.Fatal(err)
	}
	result := resultStream.First()

	if !request1.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Request1 did not arrive")
	}

	data2 := []byte{4, 5, 6, 7}
	o1.Reply(request1.Result(), data2)

	if !result.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Result 1 did not arrive")
	}

	result = resultStream.First()

	if !request2.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Request1 did not arrive")
	}

	o2.Reply(request2.Result(), data2)

	if !result.WaitUntilTimeout(100 * time.Millisecond) {
		t.Fatal("Result 2 did not arrive")
	}

}
