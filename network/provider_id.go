package network

type ProviderId int

const (
	ZMQ ProviderId = iota
	NANOMSG
)
