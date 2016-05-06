package tracker

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/service"
)

type Meta struct {
	Adport    int
	Exporting bool
	Tag       string
}

func NewMeta(adport int, cfg *config.Config, limit int) (m *Meta) {
	m = &Meta{
		Adport:    adport,
		Exporting: cfg.Exporting(),
	}

	for k, v := range cfg.Tags() {
		m.Tag = fmt.Sprintf("%s:%s", k, v)
		if len(m.Tag)+4 <= limit {
			break
		} else {
			m.Tag = ""
		}
	}
	return
}

func DecodeMeta(b []byte) (m *Meta, err error) {
	if len(b) < 4 {
		err = errors.New("Invalid Meta")
	}

	if b[0] != service.PROTOCOLL_SIGNATURE {
		err = errors.New("Invalid Meta")
	}

	bp := b[1:3]
	port := int(binary.LittleEndian.Uint16(bp))
	exporting := b[3] == 1
	tag := string(b[4:])

	m = &Meta{port, exporting, tag}
	return
}

func (m *Meta) ToBytes() (b []byte) {

	b = []byte{service.PROTOCOLL_SIGNATURE}

	bp := port2byte(m.Adport)
	b = append(b, bp[0])
	b = append(b, bp[1])

	var be byte = 0
	if m.Exporting {
		be = 1
	}
	b = append(b, be)

	for _, s := range []byte(m.Tag) {
		b = append(b, s)
	}

	return
}

func (m *Meta) TagKeyValue() (k, v string, err error) {
	s := strings.Split(m.Tag, ":")

	if len(s) == 2 {
		k = s[0]
		v = s[1]
	} else {
		err = errors.New("invalid tag")
	}

	return
}
