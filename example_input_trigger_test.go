package thingiverseio_test

import (
	"log"

	"github.com/ThingiverseIO/thingiverseio"
)

const descriptor = `
func SayHello(Greeting string) (Answer string)
tag example_tag
`

// SayHelloInput represents the input parameters for the  SayHello function.
type SayHelloInput struct {
	Greeting string
}

// SayHelloOutput represents the output parameters for the  SayHello function.
type SayHelloOutput struct {
	Answer string
}

// ExampleInputTrigger demponstrate a simple input using the TRIGGER mechanism.
func Example_inputTrigger() {
	// Create and run the input.
	i := thingiverseio.NewInput(desc)
	i.Run()

	// Start listen to the SyHello function and get a channel to receive results.
	i.StartListen("SayHello")
	c := i.ListenResults().AsChan()

	// Wait until an output connects.
	i.Connected().WaitUntilComplete()

	// Create the request parameter.
	p := SayHelloInput{"Greetings, this is a CALL example"}

	// Do the trigger.
	c := i.Trigger("SayHello", p)

	// Receive the result.
	result := <-c

	// Decode and print the result.
	var out SayHelloOutput
	result.Decode(&out)

	log.Println("Received an answer:", out.Answer)
}
