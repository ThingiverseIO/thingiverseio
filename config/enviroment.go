package config

import (
	"os"
	"strings"
)

const (
	ENV_INTERFACE = "THINGIVERSEIO_INTERFACE"
	ENV_LOGGING    = "THINGIVERSEIO_LOGGING"
	ENV_DEBUG      = "THINGIVERSEIO_DEBUG"
)

func CheckEnviroment(cfg *UserConfig) {
	CheckEnviromentLogger(cfg)
	CheckEnviromentDebug(cfg)
	CheckEnviromentInterfaces(cfg)
}

func CheckEnviromentInterfaces(cfg *UserConfig) {

	v, f := getVar(ENV_INTERFACE)
	if f {
		cfg.Interfaces = strings.Split(v, ";")
	}

}

func CheckEnviromentLogger(cfg *UserConfig) {

	v, f := getVar(ENV_LOGGING)

	if f {
		setLoggerFromString(v, cfg)
	}

}

func CheckEnviromentDebug(cfg *UserConfig) {
	v, f := getVar(ENV_DEBUG)

	if f {
		switch v {
		case "1", "true":
			cfg.Debug = true
		default:
		}
	}
}

func getVar(key string) (v string, f bool) {
	v = os.Getenv(key)
	f = v != ""
	return
}
