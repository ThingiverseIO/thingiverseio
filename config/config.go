package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/nu7hatch/gouuid"
)

type Config struct {
	debug bool

	logger io.Writer

	exporting bool

	interfaces []string //the network interfaces to use

	functionTags map[string]string
	userTags     map[string]string

	uuid UUID
}

func New(exporting bool, functionTags map[string]string) (cfg *Config) {
	cfg = &Config{
		logger:       ioutil.Discard,
		exporting:    exporting,
		interfaces:   []string{"127.0.0.1"},
		functionTags: functionTags,
		userTags:     map[string]string{},
	}
	return
}

func (cfg *Config) UUID() UUID {
	if cfg.uuid == "" {
		id, _ := uuid.NewV4()
		cfg.uuid = UUID(id.String())
	}
	return cfg.uuid
}

func (cfg *Config) Tags() (tags map[string]string) {
	tags = map[string]string{}
	for k, v := range cfg.userTags {
		tags[k] = v
	}

	for k, v := range cfg.functionTags {
		tags[k] = v
	}

	return
}

func (cfg *Config) AddOrSetUserTag(k, v string) {
	cfg.userTags[k] = v
	return
}

func (cfg *Config) Interfaces() []string {
	return cfg.interfaces
}

func (cfg *Config) OverrideInterfaces(interfaces []string) {
	cfg.interfaces = interfaces
}

func (cfg *Config) OverrideUUID(uuid string) {
	cfg.uuid = UUID(uuid)
}

func (cfg *Config) Exporting() bool {
	return cfg.exporting
}

func (cfg *Config) Logger() io.Writer {
	return cfg.logger
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
