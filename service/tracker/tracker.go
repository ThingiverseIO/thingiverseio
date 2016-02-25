package tracker

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/joernweissenborn/thingiverse.io/service"
	"github.com/joernweissenborn/thingiverse.io/service/tracker/beacon"
)

type Tracker struct {
	beacon     *beacon.Beacon
	cfg        *config.Config
	evtHandler *eventHandler
	iface      string
	logger     *log.Logger
	memberlist *memberlist.Memberlist
	port       int //memberlist bind port
	adport     int //port to advertise
}

func New(iface string, adport int, cfg *config.Config) (t *Tracker, err error) {
	t = &Tracker{
		logger:     log.New(cfg.Logger(), fmt.Sprintf("%s TRACKER ", cfg.UUID()), 0),
		cfg:        cfg,
		iface:      iface,
		adport:     adport,
		evtHandler: newEventHandler(),
	}

	t.port, err = getRandomPort(t.iface)

	if err != nil {
		return
	}

	err = t.setupMemberlist()

	return
}

func (t *Tracker) Shutdown(eventual2go.Data) (err error) {
	t.logger.Println("Stopping")
	t.evtHandler.close()
	t.StopAutoJoin()
	t.memberlist.Leave(1 * time.Second)
	err = t.memberlist.Shutdown()
	t.logger.Println("Stopped")
	return
}

func (t *Tracker) StartAutoJoin() (err error) {
	if t.beacon != nil {
		t.logger.Println("Autodiscovery already running")
		return
	}
	t.logger.Println("Starting to advertise port", t.port)

	conf := &beacon.Config{
		Addr:         t.iface,
		Port:         5557,
		PingInterval: 1 * time.Second,
		Payload:      newSignalPayload(t.port),
		Logger:       ioutil.Discard,
	}

	t.beacon, err = beacon.New(conf)
	if err != nil {
		return
	}

	t.Join().First().Then(t.silenceOnFirstJoin)
	t.beacon.Signals().Where(validSignal).Listen(t.joinOnSignal)

	t.beacon.Run()
	t.beacon.Ping()

	return
}

func (t *Tracker) StopAutoJoin() {
	if t.beacon == nil {
		return
	}
	t.beacon.Stop()
	t.beacon = nil
}

func (t *Tracker) silenceOnFirstJoin(n Node) Node {
	t.logger.Println("Joined memberlist cluster")
	t.beacon.Silence()
	return n
}
func (t *Tracker) joinOnSignal(s beacon.Signal) {
	addr := net.IP(s.SenderIp)
	port := binary.LittleEndian.Uint16(s.Data[1:])
	t.logger.Printf("Joining service %s:%d", addr, port)
	err := t.JoinCluster([]string{fmt.Sprintf("%s:%d", addr, port)})
	if err != nil {
		t.logger.Println("ERROR", err)
	}
}

func (t *Tracker) Port() int {
	return t.port
}

func (t *Tracker) Join() *NodeStream {
	return t.evtHandler.Join()
}

func (t *Tracker) Leave() *NodeStream {
	return t.evtHandler.Leave()
}

func (t *Tracker) JoinCluster(addr []string) (err error) {
	_, err = t.memberlist.Join(addr)
	return
}

func (t *Tracker) setupMemberlist() (err error) {

	conf := memberlist.DefaultLANConfig()
	conf.LogOutput = ioutil.Discard

	conf.Name = fmt.Sprintf("%s:%s", t.cfg.UUID().FullString(), t.iface)

	conf.BindAddr = t.iface
	conf.BindPort = t.port

	conf.Delegate = newDelegate(t.adport, t.cfg)
	conf.Events = t.evtHandler

	t.memberlist, err = memberlist.Create(conf)

	return
}

func getRandomPort(iface string) (int, error) {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:0", iface))
	if err != nil {
		return -1, err
	}
	defer l.Close()
	return int(l.Addr().(*net.TCPAddr).Port), nil
}

func port2byte(port int) (b []byte) {
	b = make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(port))
	return
}

func newSignalPayload(port int) (payload []byte) {
	b := port2byte(port)
	payload = []byte{service.PROTOCOLL_SIGNATURE, b[0], b[1]}
	return
}

func validSignal(s beacon.Signal) bool {
	return len(s.Data) == 3 && s.Data[0] == service.PROTOCOLL_SIGNATURE
}
