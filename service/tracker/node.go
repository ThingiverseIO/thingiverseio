package tracker

import (
	"strings"

	"github.com/hashicorp/memberlist"
	"github.com/joernweissenborn/thingiverseio/config"
)

//go:generate event_generator -t Node

type Node struct {
	*memberlist.Node
}

func (n Node) Meta() (*Meta, error) {
	return DecodeMeta(n.Node.Meta)
}

func (n Node) UUID() config.UUID {
	return config.UUID(strings.Split(n.Name, ":")[0])
}
