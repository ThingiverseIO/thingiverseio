package network

type TransportID int

const (
	ZMQ TransportID = iota
	NANOMSG
)
