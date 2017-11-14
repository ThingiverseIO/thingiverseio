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

type Transports struct {
	Details        []Details
	EncodedDetails [][]byte
	Transport       map[TransportID]Transport
	packages       *PackageStreamController
}

func NewTransports(cfg *config.Config, provider []Transport) (ps Transports, err error) {
	ps = Transports{
		Transport: map[TransportID]Transport{},
		packages: NewPackageStreamController(),
	}

	for _, p := range provider {
		if err = p.Init(cfg); err != nil {
			return
		}

		ps.Transport[p.Details().Transport] = p

		ps.packages.Join(p.Packages())

		var dtl bytes.Buffer
		enc := codec.NewEncoder(&dtl, &mh)
		enc.Encode(p.Details())

		ps.EncodedDetails = append(ps.EncodedDetails, dtl.Bytes())
		ps.Details = append(ps.Details, p.Details())
	}

	return
}

func (p Transports) Connect(details [][]byte, uuid uuid.UUID) (conn Connection, err error) {
	for _, encDetail := range details {
		var d Details
		dec := codec.NewDecoder(bytes.NewBuffer(encDetail), &mh)
		dec.Decode(&d)
		if p, ok := p.Transport[d.Transport]; ok {
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

func (p Transports) Packages() *PackageStream {
	return p.packages.Stream()
}

func (p Transports) RegisterShutdown(s *eventual2go.Shutdown) {
	for _, provider := range p.Transport {
		s.Register(provider)
	}
}
