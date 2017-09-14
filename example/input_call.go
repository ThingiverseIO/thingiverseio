
package main

import (
	"log"

	"github.com/ThingiverseIO/thingiverseio"
)

const descriptor = `
function SayHello(Greeting string) (Answer string)
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

func main() {
	// Create and run the input.
	i, err := thingiverseio.NewInput(descriptor)
	if err != nil {
		log.Fatal(err)
	}

	i.Run()

	// Create the request parameter.
	p := SayHelloInput{"Greetings, this is a CALL example"}

	// Do the call and get a channel for receiving the result
	f, err := i.Call("SayHello", p)
	if err != nil {
		log.Fatal(err)
	}
	c := f.AsChan()

	// Receive the result.
	result := <-c

	// Decode and print the result.
	var out SayHelloOutput
	result.Decode(&out)

	log.Println("Received an answer:", out.Answer)
}
