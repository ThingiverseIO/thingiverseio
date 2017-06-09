package core

import (
	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/joernweissenborn/eventual2go"
)

type properties map[string]*eventual2go.Observable

func newProperties(desc descriptor.Descriptor) (ps properties) {
	ps = properties{}
	for _, p := range desc.Properties {
		ps[p.Name] = eventual2go.NewObservable([]byte{})
	}
	return
}
