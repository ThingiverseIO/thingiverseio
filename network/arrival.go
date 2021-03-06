package network

import (
	"bytes"

	"github.com/ThingiverseIO/uuid"
	"github.com/ugorji/go/codec"
)

//go:generate evt2gogen -t Arrival

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
			if is = d.Transport == ad.Transport; is {
				return
			}
		}
	}
	return
}
