package beacon

import "time"

type Config struct {
	Address      string
	Port         int
	Payload      []byte
	PingInterval time.Duration
}
