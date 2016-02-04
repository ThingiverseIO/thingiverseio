package tracking

import (
	"fmt"
	"log"
	"net"

	"github.com/hashicorp/memberlist"
	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/thingiverse.io/config"
)

type Tracker struct {
	memberlist *memberlist.Memberlist

	logger *log.Logger

	cfg *config.Config

	iface string
	port  int

	evtHandler eventHandler
}

func Create(iface string, cfg *config.Config) (t *Tracker, err error) {
	t = &Tracker{
		logger:     log.New(cfg.Logger(), "TRACKER ", log.Ltime),
		cfg:        cfg,
		iface:      iface,
		evtHandler: newEventHandler(),
	}

	t.port, err = getRandomPort(t.iface)

	if err != nil {
		return
	}

	err = t.setupMemberlist()

	return
}

func (t *Tracker) Port() int {
	return t.port
}

func (t *Tracker) Join() *eventual2go.Stream {
	return t.evtHandler.Join()
}

func (t *Tracker) Leave() *eventual2go.Stream {
	return t.evtHandler.Leave()
}

func (t *Tracker) joinCluster(addr []string) (err error) {
	_, err = t.memberlist.Join(addr)
	return
}

func (t *Tracker) setupMemberlist() (err error) {

	conf := memberlist.DefaultLANConfig()

	conf.Name = fmt.Sprintf("%s:%s", t.cfg.UUID(), t.iface)

	conf.BindAddr = t.iface
	conf.BindPort = t.port

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
