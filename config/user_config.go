package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ThingiverseIO/thingiverseio/descriptor"
)

// UserConfig stores user configurable parameters
type UserConfig struct {
	Debug     bool   // debugging switch [false]
	Interface string // the network interface to use [1.27.0.0.1]
	Logger    string // logger [none]
	Tags      descriptor.Tagset
}

// DefaultLocalhost provides a standart config for testing purposes. Debug is true, logging set to stdout and the interface is set to '127.0.0.1'
func DefaultLocalhost() (cfg *UserConfig) {
	cfg = &UserConfig{
		Debug:     true,
		Logger:    "stdout",
		Interface: "127.0.0.1",
	}
	return
}

// GetLogger returns the configured writer for logging as io.Writer
func (cfg *UserConfig) GetLogger() (logger io.Writer) {

	switch strings.ToLower(cfg.Logger) {
	case "stdout":
		logger = os.Stdout
	case "stderr":
		logger = os.Stderr
	case "none", "":
		logger = ioutil.Discard
	default:
		_, err := os.Stat(cfg.Logger)
		if err == nil {
			logger, err = os.OpenFile(cfg.Logger, os.O_RDWR, 0666)
		} else if os.IsNotExist(err) {
			logger, err = os.Create(cfg.Logger)
			if err != nil {
				panic(fmt.Sprint("Error opening logfile: ", err))
			}
		}
	}
	return
}

func (cfg *UserConfig) String() string {
	return fmt.Sprintf(`Interfaces: %v
Logger: %s
Tags: %s
Debug: %v
`, cfg.Interface, cfg.Logger, cfg.Tags, cfg.Debug)
}
