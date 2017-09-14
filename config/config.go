// Package config provides configuration facilities. It relies heavily on github,com/spf13/viper.
//
// There are 2 types of configuration data: internal and user configuration. Internal configuration is generated automatically not intended to be changed by user. User configuration allows the user to specify the network interface to use, as well as logging output and setting the debug flag.
package config

// Config holds an internal configuration and a user configuration. Whereas the former is generated automatically, the latter is provided by the user. See function 'Configure' for details how the Userconfig is generated.
type Config struct {
	Internal *InternalConfig
	User     *UserConfig
}
