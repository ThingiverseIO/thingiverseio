package config

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func setLoggerFromString(v string, cfg *Config, logger *log.Logger) {
	switch strings.ToLower(v) {
	case "stdout":
		cfg.logger = os.Stdout
		logger.Println("Setting log to", v)
	case "stderr":
		cfg.logger = os.Stderr
		logger.Println("Setting log to", v)
	case "none":
		cfg.logger = ioutil.Discard
		logger.Println("Setting log to", v)
	case "":
	default:
		_, err := os.Stat(v)
		if err == nil {
			cfg.logger, err = os.OpenFile(v, os.O_RDWR, 0666)
			logger.Println("Setting log to", v)
			if err != nil {
				logger.Fatal(err)
			}
		} else if os.IsNotExist(err) {
			cfg.logger, err = os.Create(v)
			logger.Println("Setting log to", v)
			if err != nil {
				logger.Fatal(err)
			}
		}
	}

}
