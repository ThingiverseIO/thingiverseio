// +build !test

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

// ExampleInputTrigger demonstrates a simple input using the TRIGGER mechanism.
func Example_inputTrigger() {
	// Create and run the input.
	i, err := thingiverseio.NewInput(descriptor)
	if err != nil {
		log.Fatal(err)
	}
	i.Run()

	// Start listen to the SyHello function and get a channel to receive results.
	i.StartListen("SayHello")
	c, _ := i.ListenResults().AsChan()

	// Wait until an output connects.
	i.ConnectedFuture().WaitUntilComplete()

	// Create the request parameter.
	p := SayHelloInput{"Greetings, this is a CALL example"}

	// Do the trigger.
	i.Trigger("SayHello", p)

	// Receive the result.
	result := <-c

	// Decode and print the result.
	var out SayHelloOutput
	result.Decode(&out)

	log.Println("Received an answer:", out.Answer)
}
