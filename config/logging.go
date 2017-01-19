package config

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func setLoggerFromString(v string, cfg *UserConfig) {
	switch strings.ToLower(v) {
	case "stdout":
		cfg.Logger = os.Stdout
	case "stderr":
		cfg.Logger = os.Stderr
	case "none", "":
		cfg.Logger = ioutil.Discard
	default:
		_, err := os.Stat(v)
		if err == nil {
			cfg.Logger, err = os.OpenFile(v, os.O_RDWR, 0666)
			if err != nil {
				log.Fatal("Could not initialize logfile", err)
			}
		} else if os.IsNotExist(err) {
			cfg.Logger, err = os.Create(v)
			if err != nil {
				log.Fatal("Could not initialize logfile", err)
			}
		}
	}
}
