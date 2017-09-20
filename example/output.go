package main

import (
	"fmt"
	"log"

	"github.com/ThingiverseIO/thingiverseio"
)

const descriptor = `
function SayHello(Greeting string) (Answer string)
property Mood: Feeling string
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

type Mood struct {
	Feeling string
}

func main() {

	// Create the output.
	o, err := thingiverseio.NewOutput(descriptor)
	if err != nil {
		log.Fatal(err)
	}

	// Get a channel to receive requests
	rc, _ := o.Requests().AsChan()

	// Run the output.
	o.Run()

	o.SetProperty("Mood", Mood{"Great"})

	// Answer the requests.
	for request := range rc {
		// Print the request function.
		log.Println("Got function call:", request.Function)

		// Decode and print the input parameters.
		var in SayHelloInput
		request.Decode(&in)
		log.Println("Got greeting:", in.Greeting)

		// Reply
		out := SayHelloOutput{fmt.Sprint("Greetings back, you said:", in.Greeting)}
		o.Reply(request, out)
	}
}
