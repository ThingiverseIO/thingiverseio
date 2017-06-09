package thingiverseio

import "github.com/joernweissenborn/eventual2go"

//go:generate event_generator -t Property

type Property struct {
	Name  string
	value []byte
}

func (p Property) Value(value interface{}) (err error) {
	err = decode(p.value, value)
	return
}

func propertyFromFuture(property string) eventual2go.CompletionHandler {
	return func(d eventual2go.Data) eventual2go.Data {
		return Property{
			Name:  property,
			value: d.([]byte),
		}
	}
}
