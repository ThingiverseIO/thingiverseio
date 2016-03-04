package config

import (
	"io"

)

func Configure(logger io.Writer, exporting bool, functionTags map[string]string) (cfg *Config) {

	cfg = New(logger, exporting, functionTags)

	CheckCfgFile(cfg, CfgFileGlobal())

	CheckEnviroment(cfg)

	CheckCfgFile(cfg, CfgFileUser())

	return
}
