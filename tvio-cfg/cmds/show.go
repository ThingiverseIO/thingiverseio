package cmds

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/mitchellh/cli"
)

func ShowCommandFactory() (cli.Command, error) {
	return &ShowCommand{}, nil
}

type ShowCommand struct{}

func (*ShowCommand) Help() string {
	return `Shows the current configuration of the thingiverse.io network on this machine`
}
func (*ShowCommand) Synopsis() string {
	return `Shows the current configuration of the thingiverse.io network on this machine`
}

func (*ShowCommand) Run(args []string) int {

	var logger = log.New(os.Stdout, "", 0)
	checklogger := ioutil.Discard

	cfg := config.New(checklogger)

	config.CheckEnviroment(cfg)

	logger.Printf(`
Enviroment Configuration
========================

%s
`, cfg)

	//Global Dir

	cfg = config.New(checklogger)
	logger.Printf(`
Global Configuration
========================
file: %s

`, config.CfgFileGlobal())

	if config.CfgFileGlobalPresent() {
		config.CheckCfgFile(cfg, config.CfgFileGlobal())
		logger.Println(cfg)
	} else {
		logger.Println("Not Present")
	}

	//User Dir

	cfg = config.New(checklogger)
	logger.Printf(`
User Configuration
========================
file: %s

`, config.CfgFileUser())

	if config.CfgFileUserPresent() {
		config.CheckCfgFile(cfg, config.CfgFileUser())
		logger.Println(cfg)
	} else {
		logger.Println("Not Present")
	}

	//WD

	cfg = config.New(checklogger)
	logger.Printf(`
Working Dir Configuration
========================
file: %s

`, config.CfgFileCwd())

	if config.CfgFileCwdPresent() {
		config.CheckCfgFile(cfg, config.CfgFileCwd())
		logger.Println(cfg)
	} else {
		logger.Println("Not Present")
	}

	logger.Printf(`
Configuration Used
==================

%s
`, config.Configuration)

	return 0
}
