package network

import (
	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/joernweissenborn/eventual2go"
)

type MockProvider struct {
	details      Details
	messageBoxes []*MessageStreamController
	cfg          *config.Config
}

func NewMockProvider(nr int) (p []*MockProvider) {
	messageBoxes := []*MessageStreamController{}

	for i := 0; i < nr; i++ {
		messageBoxes = append(messageBoxes, NewMessageStreamController())
	}
	for i := 0; i < nr; i++ {
		p = append(p, &MockProvider{
			messageBoxes: messageBoxes,
			details:      Details{Provider: 0, Config: []byte{byte(i)}},
		})

	}
	return
}

func (m *MockProvider) Init(cfg *config.Config) error {
	m.cfg = cfg
	return nil
}

func (m *MockProvider) Connect(details Details, uuid uuid.UUID) (Connection, error) {
	ams, _ := eventual2go.SpawnActor(&MockConnection{
		cfg:        m.cfg,
		messageBox: m.messageBoxes[details.Config[0]],
	})
	return Connection{ams, uuid}, nil
}

func (m *MockProvider) Details() Details {
	return m.details
}

func (m *MockProvider) Messages() *MessageStream {
	return m.messageBoxes[m.details.Config[0]].Stream()
}

func (m *MockProvider) Shutdown(d eventual2go.Data) error { return nil }

type MockConnection struct {
	cfg        *config.Config
	messageBox *MessageStreamController
}

func (m *MockConnection) Init() error {
	return nil
}

func (m *MockConnection) OnMessage(d eventual2go.Data) {
	msg := d.(Message)
	msg.Sender = m.cfg.Internal.UUID
	m.messageBox.Add(msg)
}

func (m *MockConnection) Shutdown(d eventual2go.Data) {}

type MockTracker struct {
	Av *ArrivalStreamController
	Lv *uuid.UUIDStreamController
	Dt [][]byte
}

func (m *MockTracker) Init(cfg *config.Config, dt [][]byte) error {
	m.Av = NewArrivalStreamController()
	m.Lv = uuid.NewUUIDStreamController()
	m.Dt = dt
	return nil
}

func (m *MockTracker) Arrivals() *ArrivalStream {
	return m.Av.Stream()
}

func (m *MockTracker) Leaving() *uuid.UUIDStream {
	return m.Lv.Stream()
}

func (m *MockTracker) Shutdown(d eventual2go.Data) error { return nil }
