package config

import (
	"os"
	"strings"
)

const (
	ENV_INTERFACES = "THINGIVERSEIO_INTERFACE"
	ENV_LOGGING    = "THINGIVERSEIO_LOGGING"
	ENV_DEBUG      = "THINGIVERSEIO_DEBUG"
)

func CheckEnviroment(cfg *Config) {
	CheckEnviromentLogger(cfg)
	CheckEnviromentDebug(cfg)
	CheckEnviromentInterfaces(cfg)
}

func CheckEnviromentInterfaces(cfg *Config) {

	v, f := getVar(ENV_INTERFACES)
	if f {
		cfg.interfaces = strings.Split(v, ";")
	}

}

func CheckEnviromentLogger(cfg *Config) {

	v, f := getVar(ENV_LOGGING)

	if f {
		setLoggerFromString(v, cfg)
	}

}

func CheckEnviromentDebug(cfg *Config) {
	v, f := getVar(ENV_DEBUG)

	if f {
		switch v {
		case "1", "true":
			cfg.debug = true
		default:
		}
	}
}

func getVar(key string) (v string, f bool) {
	v = os.Getenv(key)
	f = v != ""
	return
}
