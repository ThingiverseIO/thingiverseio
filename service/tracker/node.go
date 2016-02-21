package tracker

import "github.com/hashicorp/memberlist"

//go:generate event_generator -t Node

type Node struct {
	*memberlist.Node
}

func (n Node) Meta() (*Meta, error) {
	return DecodeMeta(n.Node.Meta)
}
