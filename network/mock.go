package network

import (
	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/uuid"
	"github.com/joernweissenborn/eventual2go"
)

type MockTransport struct {
	details      Details
	messageBoxes []*PackageStreamController
	cfg          *config.Config
}

func NewMockTransport(nr int) (p []*MockTransport) {
	messageBoxes := []*PackageStreamController{}

	for i := 0; i < nr; i++ {
		messageBoxes = append(messageBoxes, NewPackageStreamController())
	}
	for i := 0; i < nr; i++ {
		p = append(p, &MockTransport{
			messageBoxes: messageBoxes,
			details:      Details{Transport: 0, Config: []byte{byte(i)}},
		})

	}
	return
}

func (m *MockTransport) Init(cfg *config.Config) error {
	m.cfg = cfg
	return nil
}

func (m *MockTransport) Connect(details Details, uuid uuid.UUID) (Connection, error) {
	ams, _ := eventual2go.SpawnActor(&MockConnection{
		cfg:        m.cfg,
		messageBox: m.messageBoxes[details.Config[0]],
	})
	return Connection{ams, uuid}, nil
}

func (m *MockTransport) Details() Details {
	return m.details
}

func (m *MockTransport) Packages() *PackageStream {
	return m.messageBoxes[m.details.Config[0]].Stream()
}

func (m *MockTransport) Shutdown(d eventual2go.Data) error { return nil }

type MockConnection struct {
	cfg        *config.Config
	messageBox *PackageStreamController
}

func (m *MockConnection) Init() error {
	return nil
}

func (m *MockConnection) OnMessage(d eventual2go.Data) {
	msg := d.(Package)
	msg.Sender = m.cfg.Internal.UUID
	m.messageBox.Add(msg)
}

func (m *MockConnection) Shutdown(d eventual2go.Data) error { return nil }

type MockTracker struct {
	Av      *ArrivalStreamController
	Lv      *uuid.UUIDStreamController
	Dt      [][]byte
	Partner *ArrivalStreamController
	UUID    uuid.UUID
}

func (m *MockTracker) Init(cfg *config.Config, dt [][]byte) error {
	if m.Av == nil {
		m.Av = NewArrivalStreamController()
	}
	m.Lv = uuid.NewUUIDStreamController()
	m.Dt = dt
	m.UUID = cfg.Internal.UUID
	return nil
}

func (m *MockTracker) Arrivals() *ArrivalStream {
	return m.Av.Stream()
}

func (m *MockTracker) Leaving() *uuid.UUIDStream {
	return m.Lv.Stream()
}

func (m *MockTracker) StartAdvertisment() error {
	if m.Partner != nil {
		m.Partner.Add(Arrival{
			Details: m.Dt,
			UUID:    m.UUID,
		})
	}
	return nil
}

func (m *MockTracker) StopAdvertisment() {}

func (m *MockTracker) Shutdown(d eventual2go.Data) error { return nil }
