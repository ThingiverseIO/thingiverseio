package config

import (
	"io/ioutil"
	"log"
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

	logger := log.New(cfg.logger, "CONFIG_FILE_CHECK ", log.Ltime)
	logger.Println("Checking File", path)
	f, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Println("Error reading file", err)
	}

	var cfgf CfgFile

	err = gcfg.ReadStringInto(&cfgf, string(f))

	if err != nil {
		logger.Println("Error parsing file", err)
	}

	if len(cfgf.Network.Interface) != 0 {
		cfg.interfaces = cfgf.Network.Interface
		logger.Println("Setting interfaces to", cfg.interfaces)
	}
	setLoggerFromString(cfgf.Misc.Logging, cfg, logger)
	parseUserTags(cfgf.UserTags.Tag, logger, cfg)
}

func parseUserTags(t []string, logger *log.Logger, cfg *Config) {
	for _, ut := range t {
		if !strings.Contains(ut, ":") {
			logger.Printf("Error Parsing Tag '%s', tags must be of form 'key:value'", ut)
			continue
		}
		split := strings.Split(ut, ":")
		if len(split) != 2 {
			logger.Printf("Error Parsing Tag '%s', tags must be of form 'key:value'", ut)
			continue
		}
		cfg.userTags[split[0]] = split[1]
	}
}
