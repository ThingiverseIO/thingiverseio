package beacon

import "time"

type Config struct {
	Addr         string
	Port         int
	Payload      []byte
	PingInterval time.Duration
}
