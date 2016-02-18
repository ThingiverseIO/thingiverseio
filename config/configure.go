package config

import (
	"io"

)

func Configure(logger io.Writer, exporting bool) (cfg *Config) {

	cfg = New(logger, exporting)

	CheckCfgFile(cfg, CfgFileGlobal())

	CheckEnviroment(cfg)

	CheckCfgFile(cfg, CfgFileUser())

	return
}
