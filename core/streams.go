package core

import (
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/joernweissenborn/eventual2go"
)

type streams map[string]*eventual2go.StreamController

func newStreams(desc descriptor.Descriptor) (ss streams) {
	ss = streams{}
	for _, s := range desc.Streams {
		ss[s.Name] = eventual2go.NewStreamController()
	}
	return
}
