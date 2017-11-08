package network

import (
	"bytes"
	"errors"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/uuid"
	"github.com/joernweissenborn/eventual2go"
	"github.com/ugorji/go/codec"
)

var mh codec.MsgpackHandle

type Providers struct {
	Details        []Details
	EncodedDetails [][]byte
	Provider       map[ProviderID]Provider
	messages       *MessageStreamController
}

func NewProviders(cfg *config.Config, provider []Provider) (ps Providers, err error) {
	ps = Providers{
		Provider: map[ProviderID]Provider{},
		messages: NewMessageStreamController(),
	}

	for _, p := range provider {
		if err = p.Init(cfg); err != nil {
			return
		}

		ps.Provider[p.Details().Provider] = p

		ps.messages.Join(p.Messages())

		var dtl bytes.Buffer
		enc := codec.NewEncoder(&dtl, &mh)
		enc.Encode(p.Details())

		ps.EncodedDetails = append(ps.EncodedDetails, dtl.Bytes())
		ps.Details = append(ps.Details, p.Details())
	}

	return
}

func (p Providers) Connect(details [][]byte, uuid uuid.UUID) (conn Connection, err error) {
	for _, encDetail := range details {
		var d Details
		dec := codec.NewDecoder(bytes.NewBuffer(encDetail), &mh)
		dec.Decode(&d)
		if p, ok := p.Provider[d.Provider]; ok {
			conn, err = p.Connect(d, uuid)
			if err == nil {
				return
			}
		}
	}
	if err == nil {
		err = errors.New("No provider found.")
	}
	return
}

func (p Providers) Messages() *MessageStream {
	return p.messages.Stream()
}

func (p Providers) RegisterShutdown(s *eventual2go.Shutdown) {
	for _, provider := range p.Provider {
		s.Register(provider)
	}
}
