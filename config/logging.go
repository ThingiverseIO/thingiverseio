package config

import (
	"io/ioutil"
	"os"
	"strings"
)

func setLoggerFromString(v string, cfg *Config) {
	switch strings.ToLower(v) {
	case "stdout":
		cfg.logger = os.Stdout
	case "stderr":
		cfg.logger = os.Stderr
	case "none", "":
		cfg.logger = ioutil.Discard
	default:
		_, err := os.Stat(v)
		if err == nil {
			cfg.logger, err = os.OpenFile(v, os.O_RDWR, 0666)
			if err != nil {
				panic(err)
			}
		} else if os.IsNotExist(err) {
			cfg.logger, err = os.Create(v)
			if err != nil {
				panic(err)
			}
		}
	}
}
