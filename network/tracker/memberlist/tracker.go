package memberlist

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/network"
	"github.com/ThingiverseIO/thingiverseio/network/tracker/beacon"
	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/hashicorp/memberlist"
	"github.com/joernweissenborn/eventual2go"
	"github.com/ugorji/go/codec"
)

var (
	mh codec.MsgpackHandle
)

type Tracker struct {
	arrivals   *network.ArrivalStreamController
	leaving    *uuid.UUIDStreamController
	beacon     *beacon.Beacon
	memberlist *memberlist.Memberlist
	meta       []byte
}

// Init initializes the tracker.
func (t *Tracker) Init(cfg *config.Config, details [][]byte) (err error) {

	t.arrivals = network.NewArrivalStreamController()
	t.leaving = uuid.NewUUIDStreamController()

	var buf bytes.Buffer
	enc := codec.NewEncoder(&buf, &mh)
	enc.Encode(network.Arrival{
		IsOutput: cfg.Internal.Output,
		Details:  details,
		UUID:     cfg.Internal.UUID,
	})
	t.meta = buf.Bytes()

	var port int
	if port, err = t.setupMemberlist(cfg.User.Interfaces[0], cfg.Internal.UUID); err != nil {
		return
	}

	err = t.setupBeacon(cfg.User.Interfaces[0], port)
	return
}

// Arrivals return a stream of arrived peers.
func (t *Tracker) Arrivals() *network.ArrivalStream {
	return t.arrivals.Stream()
}

// Leaving returns a stream of leaving peers.
func (t *Tracker) Leaving() *uuid.UUIDStream {
	return t.leaving.Stream()
}

func (t *Tracker) StartAdvertisment() (err error) {
	t.beacon.Ping()
	return
}

func (t *Tracker) StopAdvertisment() {
	t.beacon.Silence()
}

func (t *Tracker) Shutdown(eventual2go.Data) (err error) {
	t.beacon.Silence()
	t.beacon.Stop()
	err = t.memberlist.Shutdown()
	return
}

func (t *Tracker) NotifyJoin(n *memberlist.Node) {

	dec := codec.NewDecoder(bytes.NewBuffer(n.Meta), &mh)
	var arr network.Arrival
	dec.Decode(&arr)
	t.arrivals.Add(arr)
}

func (t *Tracker) NotifyLeave(n *memberlist.Node) {
	t.leaving.Add(uuid.UUID(n.Name))
}

func (t *Tracker) NotifyUpdate(n *memberlist.Node) {
	// not handled at the moment
}

func (t *Tracker) NodeMeta(limit int) (meta []byte) {
	meta = t.meta
	return
}

func (t *Tracker) NotifyMsg([]byte) {
	// not implemented
}

func (t *Tracker) GetBroadcasts(overhead, limit int) [][]byte {
	// not implemented
	return nil
}

func (t *Tracker) LocalState(join bool) []byte {
	// not implemented
	return nil
}

func (t *Tracker) MergeRemoteState(buf []byte, join bool) {
	// not implemented
}

func (t *Tracker) onSignal(s beacon.Signal) {
	addr := net.IP(s.SenderIp)
	port := binary.LittleEndian.Uint16(s.Data[1:])
	t.memberlist.Join([]string{fmt.Sprintf("%s:%d", addr, port)})
}

func (t *Tracker) setupBeacon(iface string, port int) (err error) {

	conf := &beacon.Config{
		Addr:         iface,
		Port:         5557,
		PingInterval: 1 * time.Second,
		Payload:      newSignalPayload(port),
	}

	if t.beacon, err = beacon.New(conf); err != nil {
		return
	}

	t.beacon.Signals().Where(validSignal).Listen(t.onSignal)

	t.beacon.Run()
	return
}

func (t *Tracker) setupMemberlist(iface string, uuid uuid.UUID) (port int, err error) {

	conf := memberlist.DefaultLANConfig()
	conf.LogOutput = ioutil.Discard

	conf.Name = uuid.FullString()

	if port, err = network.GetFreePortOnInterface(iface); err != nil {
		return
	}

	conf.BindAddr = iface
	conf.BindPort = port

	conf.Delegate = t
	conf.Events = t

	t.memberlist, err = memberlist.Create(conf)

	return
}

func port2byte(port int) (b []byte) {
	b = make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(port))
	return
}

func newSignalPayload(port int) (payload []byte) {
	b := port2byte(port)
	payload = []byte{network.PROTOCOLL_SIGNATURE, b[0], b[1]}
	return
}

func validSignal(s beacon.Signal) bool {
	return len(s.Data) == 3 && s.Data[0] == network.PROTOCOLL_SIGNATURE
}
