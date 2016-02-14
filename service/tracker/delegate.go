package tracker

import (
	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/joernweissenborn/thingiverse.io/service"
)

type memberlistDelegate struct {
	adport int
	cfg    *config.Config
}

func newDelegate(adport int, cfg *config.Config) (d *memberlistDelegate) {
	return &memberlistDelegate{adport, cfg}
}

func (md *memberlistDelegate) NodeMeta(limit int) (meta []byte) {

	meta = []byte{service.PROTOCOLL_SIGNATURE}

	bp := port2byte(md.adport)
	meta = append(meta, bp[0])
	meta = append(meta, bp[1])

	var b byte = 0
	if md.cfg.Exporting() {
		b = 1
	}
	meta = append(meta, b)

	var t string
	for k, v := range md.cfg.Tags() {
		t = k + v
		if len(t)+4 <= limit {
			break
		} else {
			t = ""
		}
	}

	for _, s := range []byte(t) {
		meta = append(meta, s)
	}
	return
}
func (md *memberlistDelegate) NotifyMsg([]byte) {
	// not implemented
}

func (md *memberlistDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	// not implemented
	return nil
}

func (md *memberlistDelegate) LocalState(join bool) []byte {
	// not implemented
	return nil
}

func (md *memberlistDelegate) MergeRemoteState(buf []byte, join bool) {
	// not implemented
}
