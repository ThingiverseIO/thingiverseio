package config

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/joernweissenborn/thingiverseio/config"
)

var ShowCommand = cli.Command{
	Name:        "show",
	Usage:       "Show the current configuration",
	Description: "Shows the current configuration of the thingiverse.io network on this machine",
	Action:      runShow,
}

func runShow(c *cli.Context) {

	cfg := config.New(false, map[string]string{})

	config.CheckEnviroment(cfg)

	fmt.Println(fmt.Sprintf(`
Enviroment Configuration
========================

%s
`, cfg))

	//Global Dir

	cfg = config.New(false, map[string]string{})
	fmt.Println(fmt.Sprintf(`
Global Configuration
========================
file: %s

`, config.CfgFileGlobal()))

	if config.CfgFileGlobalPresent() {
		config.CheckCfgFile(cfg, config.CfgFileGlobal())
		fmt.Println(cfg.String())
	} else {
		fmt.Println("Not Present")
	}

	//User Dir

	cfg = config.New(false, map[string]string{})

	fmt.Println(fmt.Sprintf(`
User Configuration
========================
file: %s

`, config.CfgFileUser()))

	if config.CfgFileUserPresent() {
		config.CheckCfgFile(cfg, config.CfgFileUser())
		fmt.Println(cfg.String())
	} else {
		fmt.Println("Not Present")
	}

	//WD

	cfg = config.New(false, map[string]string{})
	fmt.Println(fmt.Sprintf(`
Working Dir Configuration
========================
file: %s

`, config.CfgFileCwd()))

	if config.CfgFileCwdPresent() {
		config.CheckCfgFile(cfg, config.CfgFileCwd())
		fmt.Println(cfg.String())
	} else {
		fmt.Println("Not Present")
	}

	fmt.Println(fmt.Sprintf(`
Configuration Used
==================

%s
`, config.Configure(false, map[string]string{})))

}
