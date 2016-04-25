package config

func Configure(exporting bool, functionTags map[string]string) (cfg *Config) {

	cfg = New(exporting, functionTags)

	CheckCfgFile(cfg, CfgFileGlobal())

	CheckEnviroment(cfg)

	CheckCfgFile(cfg, CfgFileUser())

	CheckCfgFile(cfg, CfgFileCwd())

	return
}
