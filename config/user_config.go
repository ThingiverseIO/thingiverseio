package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ThingiverseIO/thingiverseio/descriptor"
)

type UserConfig struct {
	Debug      bool
	Interfaces []string //the network interfaces to use
	Logger     io.Writer
	Tags       descriptor.Tagset
}

func DefaultLocalhost() (cfg *UserConfig) {
	cfg = &UserConfig{
		Debug: true,
		Logger:     os.Stderr,
		Interfaces: []string{"127.0.0.1"},
	}
	return
}

func (cfg *UserConfig) String() string {

	lstring := ""
	switch cfg.Logger.(type) {
	case *os.File:
		lstring = cfg.Logger.(*os.File).Name()
	default:
		lstring = "none"

	}

	return fmt.Sprintf(`Interfaces: %v
Logger: %s
Tags: %s
Debug: %v
`, strings.Join(cfg.Interfaces, ", "), lstring, cfg.Tags, cfg.Debug)
}
