package config

import "io/ioutil"

var Configuration *Config

func init() {
	Configuration = New(ioutil.Discard)

	CheckEnviroment(Configuration)
	CheckCfgFile(Configuration, CfgFileGlobal())
	CheckCfgFile(Configuration, CfgFileUser())
	CheckCfgFile(Configuration, CfgFileCwd())
}
