package config

import (
	"os"
	"strings"

	"gopkg.in/gcfg.v1"
)

type CfgFile struct {
	Network  CfgFileNetwork
	Misc     CfgFileMisc
	UserTags CfgFileUserTags
}

type CfgFileNetwork struct {
	Interface []string
}

type CfgFileMisc struct {
	Logging string
	Debug   bool
}

type CfgFileUserTags struct {
	Tag []string
}

func CheckCfgFile(cfg *Config, path string) {

	cfgf, err := ReadCfgFile(path)

	if err != nil {
		return
	}

	if len(cfgf.Network.Interface) != 0 {
		cfg.interfaces = cfgf.Network.Interface
	}
	setLoggerFromString(cfgf.Misc.Logging, cfg)
	parseUserTags(cfgf.UserTags.Tag, cfg)
}

func ReadCfgFile(path string) (cfgf CfgFile, err error) {

	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	err = gcfg.ReadInto(&cfgf, f)

	return

}

func parseUserTags(t []string, cfg *Config) {
	for _, ut := range t {
		if !strings.Contains(ut, ":") {
			continue
		}
		split := strings.Split(ut, ":")
		if len(split) != 2 {
			continue
		}
		cfg.userTags[split[0]] = split[1]
	}
}
