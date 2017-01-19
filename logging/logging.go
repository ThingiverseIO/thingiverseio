package logging

import (
	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/op/go-logging"
)

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{module} %{shortfunc} ▶ %{level} %{id:03x}%{color:reset} %{message}`,
)

func SetupLogger(cfg *config.Config) {

	lvl := logging.INFO
	if cfg.User.Debug {
		lvl = logging.DEBUG
	}

	backend := logging.AddModuleLevel(
		logging.NewBackendFormatter(
			logging.NewLogBackend(cfg.User.Logger, "", 0), format))
	backend.SetLevel(lvl, "")

	logging.SetBackend(backend)
}

func GetLogger(prefix string) *logging.Logger {
	return logging.MustGetLogger(prefix)
}

func CreateLogger(prefix string, cfg *config.Config) (l *logging.Logger) {
	l = logging.MustGetLogger(prefix)
	SetupLogger(cfg)
	return
}
