package network

import (
	"bytes"

	"github.com/ThingiverseIO/uuid"
	"github.com/ugorji/go/codec"
)

//go:generate event_generator -t Arrival

type Arrival struct {
	IsOutput bool
	Details  [][]byte
	UUID     uuid.UUID
}

func (a Arrival) Supported(details []Details) (is bool) {
	for _, encAd := range a.Details {
		var ad Details
		dec := codec.NewDecoder(bytes.NewBuffer(encAd), &mh)
		dec.Decode(&ad)
		for _, d := range details {
			if is = d.Provider == ad.Provider; is {
				return
			}
		}
	}
	return
}
