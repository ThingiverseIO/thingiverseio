package config

import (
	"os/user"

	"github.com/spf13/viper"
)

var userCfg *UserConfig

//Configure loads the configuration from either the disk or the enviroment.
func Configure() (cfg *UserConfig) {

	if userCfg != nil {
		return userCfg
	}

	viper.SetDefault("debug", false)
	viper.SetDefault("logger", "none")
	viper.SetDefault("interface", "127.0.0.1")

	//Enviroment
	viper.SetEnvPrefix("tvio")
	viper.AutomaticEnv()

	//Configfile
	viper.SetConfigName(".tvio")
	viper.AddConfigPath(".") // First look in CWD

	usr, err := user.Current()
	if err != nil {
		viper.AddConfigPath(usr.HomeDir) // Then in user home
	}

	return
}
