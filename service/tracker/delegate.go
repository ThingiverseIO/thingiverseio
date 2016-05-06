package tracker

import "github.com/ThingiverseIO/thingiverseio/config"

type memberlistDelegate struct {
	adport int
	cfg    *config.Config
}

func newDelegate(adport int, cfg *config.Config) (d *memberlistDelegate) {
	return &memberlistDelegate{adport, cfg}
}

func (md *memberlistDelegate) NodeMeta(limit int) (meta []byte) {

	m := NewMeta(md.adport, md.cfg, limit)

	meta = m.ToBytes()

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
