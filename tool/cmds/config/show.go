package cmds

import (
	"fmt"
	"io/ioutil"

	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/mitchellh/cli"
)

type ShowCommand struct {
	Ui cli.Ui
}

func (*ShowCommand) Help() string {
	return `Shows the current configuration of the thingiverse.io network on this machine`
}
func (*ShowCommand) Synopsis() string {
	return `Shows the current configuration of the thingiverse.io network on this machine`
}

func (sc *ShowCommand) Run(args []string) int {

	checklogger := ioutil.Discard

	cfg := config.New(checklogger, false)

	config.CheckEnviroment(cfg)

	sc.Ui.Info(fmt.Sprintf(`
Enviroment Configuration
========================

%s
`, cfg))

	//Global Dir

	cfg = config.New(checklogger, false)
	sc.Ui.Info(fmt.Sprintf(`
Global Configuration
========================
file: %s

`, config.CfgFileGlobal()))

	if config.CfgFileGlobalPresent() {
		config.CheckCfgFile(cfg, config.CfgFileGlobal())
		sc.Ui.Info(cfg.String())
	} else {
		sc.Ui.Info("Not Present")
	}

	//User Dir

	cfg = config.New(checklogger, false)

	sc.Ui.Info(fmt.Sprintf(`
User Configuration
========================
file: %s

`, config.CfgFileUser()))

	if config.CfgFileUserPresent() {
		config.CheckCfgFile(cfg, config.CfgFileUser())
		sc.Ui.Info(cfg.String())
	} else {
		sc.Ui.Info("Not Present")
	}

	//WD

	cfg = config.New(checklogger, false)
	sc.Ui.Info(fmt.Sprintf(`
Working Dir Configuration
========================
file: %s

`, config.CfgFileCwd()))

	if config.CfgFileCwdPresent() {
		config.CheckCfgFile(cfg, config.CfgFileCwd())
		sc.Ui.Info(cfg.String())
	} else {
		sc.Ui.Info("Not Present")
	}

	sc.Ui.Info(fmt.Sprintf(`
Configuration Used
==================

%s
`, config.Configure(ioutil.Discard, false)))

	return 0
}
