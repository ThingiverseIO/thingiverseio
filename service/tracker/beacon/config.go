package beacon

import (
	"io"
	"io/ioutil"
	"time"
)

type Config struct {

	Addr string
	Port int

	PingInterval time.Duration

	Payload []byte

	Logger io.Writer
}

func (c *Config) init() {
	if c.Logger == nil {
		c.Logger = ioutil.Discard
	}
}
