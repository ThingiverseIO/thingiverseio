package config

// Config holds an internal configuration and a user configuration. Whereas the former is generated automatically, the latter is provided by the user. See function 'Configure' for details how the Userconfig is generated.
type Config struct {
	Internal *InternalConfig
	User     *UserConfig
}
