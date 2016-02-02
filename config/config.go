package config

import (
	"fmt"
	"io"
	"os"
)

type Config struct {
	debug bool

	logger io.Writer

	interfaces []string //the network interfaces to use

	userTags map[string]string
}

func New(logger io.Writer) (cfg *Config) {
	cfg = &Config{
		logger:     logger,
		interfaces: []string{"127.0.0.1"},
		userTags:   map[string]string{},
	}
	return
}

func (cfg *Config) String() string {
	istring := ""
	for i, iface := range cfg.interfaces {
		if i != 0 {
			istring += "; "
		}
		istring += iface
	}

	var lstring string

	switch cfg.logger.(type) {
	//case os.Stdout:
	//lstring = "stdout"
	//case os.Stderr:
	//lstring = "stderr"
	case *os.File:
		lstring = cfg.logger.(*os.File).Name()
	default:
		lstring = "none"

	}

	tstring := ""

	for k, v := range cfg.userTags {
		tstring += fmt.Sprintf("%s:%s; ", k, v)
	}
	if tstring != "" {
		tstring = tstring[:len(tstring)-2]
	}
	return fmt.Sprintf(`Interfaces: %v
Logger: %s
UserTags: %s
Debug: %v
`, istring, lstring, tstring, cfg.debug)
}
