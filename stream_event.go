package thingiverseio

import "github.com/joernweissenborn/eventual2go"

//go:generate evt2gogen -t StreamEvent

type StreamEvent struct {
	Name  string
	value []byte
}

func (p StreamEvent) Value(value interface{}) (err error) {
	err = decode(p.value, value)
	return
}

func toStreamEvent(stream string) eventual2go.Transformer{
	return func(d eventual2go.Data) eventual2go.Data{
		return StreamEvent{
			Name:  stream,
			value: d.([]byte),
		}
	}
}
