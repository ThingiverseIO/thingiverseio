package thingiverseio

import "github.com/joernweissenborn/eventual2go"

//go:generate evt2gogen -t Property

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

func propertyFromChange(property string, s PropertySubscriber) eventual2go.Subscriber{
	return func(d eventual2go.Data) {
		p := Property{
			Name:  property,
			value: d.([]byte),
		}
		s(p)
	}
}
