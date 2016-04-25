package config

import "github.com/codegangsta/cli"

var Command = cli.Command{
	Name:  "configure",
	Aliases: []string{"cfg"},
	Usage: "Options for configurating thingiverse.io",
	Description: "Provides tools to display the current configuration and init an empty configuration file",
	Subcommands: []cli.Command{
		InitCmd,
		ShowCommand,
	},
}
