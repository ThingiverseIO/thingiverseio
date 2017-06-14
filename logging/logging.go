package logging

import (
	"sync"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/op/go-logging"
)

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{module} ▶ %{level} %{id:03x} %{shortfunc}%{color:reset} %{message}`,
)

var m = &sync.Mutex{}

func setupLogger(cfg *config.Config) error {

	lvl := logging.INFO
	if cfg.User.Debug {
		lvl = logging.DEBUG
	}

	logger, err := cfg.User.GetLogger()
	if err != nil {
		return err
	}
	backend := logging.AddModuleLevel(
		logging.NewBackendFormatter(
			logging.NewLogBackend(logger, "", 0), format))
	backend.SetLevel(lvl, "")

	logging.SetBackend(backend)
	return nil
}

func CreateLogger(prefix string, cfg *config.Config) (l *logging.Logger) {
	m.Lock()
	defer m.Unlock()
	l = logging.MustGetLogger(prefix)
	setupLogger(cfg)
	return
}
