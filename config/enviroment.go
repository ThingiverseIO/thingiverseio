package config

import (
	"log"
	"os"
	"strings"
)

const (
	ENV_INTERFACES = "THINGIVERSEIO_INTERFACES"
	ENV_LOGGING    = "THINGIVERSEIO_LOGGING"
	ENV_DEBUG      = "THINGIVERSEIO_DEBUG"
)

func CheckEnviroment(cfg *Config) {
	logger := log.New(cfg.logger, "ENVIROMENT_CHECK ", log.Ltime)

	CheckEnviromentLogger(cfg,logger)
	CheckEnviromentDebug(cfg,logger)
	CheckEnviromentInterfaces(cfg,logger)

}


func CheckEnviromentInterfaces(cfg *Config, logger *log.Logger) {

	v, f := getVar(ENV_INTERFACES, logger)

	if f {
		cfg.interfaces = strings.Split(v, ":")
		logger.Println("Setting interfaces to", cfg.interfaces)
	}

}

func CheckEnviromentLogger(cfg *Config, logger *log.Logger) {

	v, f := getVar(ENV_LOGGING, logger)

	if f {
		setLoggerFromString(v,cfg,logger)
	}

}

func CheckEnviromentDebug(cfg *Config, logger *log.Logger) {
	v,f := getVar(ENV_DEBUG, logger)

	if f {
		switch v {
		case "1":
		logger.Println("Setting DEBUG on")
		cfg.debug = true
		default:
		logger.Println("Setting DEBUG off")
		}
	}
}

func getVar(key string, l *log.Logger) (v string, f bool) {

	l.Printf("Looking if %s is set", key)
	v = os.Getenv(key)
	f = v != ""
	if f {
		l.Println("Key found, value is", v)
	} else {
		l.Println("Key not found")
	}
	return
}
