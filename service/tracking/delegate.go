package tracking

import "github.com/joernweissenborn/thingiverse.io/config"

type memberlistDelegate struct {
	cfg *config.Config
}

func newDelegate(cfg *config.Config) (d *memberlistDelegate) {
	return &memberlistDelegate{cfg}
}

func (md *memberlistDelegate) NodeMeta(limit int) (meta []byte) {

	meta = []byte{0xA5}

	var b byte = 0
	if md.cfg.Exporting() {
		b = 1
	}
	meta = append(meta, b)

	var t string
	for k, v := range md.cfg.Tags() {
		t = k + v
		if len(t)+2 <= limit {
			break
		} else {
			t = ""
		}
	}

	for _, s := range []byte(t) {
		meta = append(meta,s)
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
