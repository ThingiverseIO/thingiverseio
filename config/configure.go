package config

func Configure() (cfg *UserConfig) {

	cfg = &UserConfig{}

	CheckCfgFile(cfg, CfgFileGlobal())

	CheckEnviroment(cfg)

	CheckCfgFile(cfg, CfgFileUser())

	CheckCfgFile(cfg, CfgFileCwd())

	return
}
