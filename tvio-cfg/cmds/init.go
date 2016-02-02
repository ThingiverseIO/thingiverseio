package cmds

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/mitchellh/cli"
)

type InitCmd struct {
	Ui cli.Ui
}

func (*InitCmd) Help() string {
	return fmt.Sprintf(`Initializes a dafault %s at the current directory. 

Specify a path or use subcommands to init a %s at different places`, config.CFG_FILE_NAME, config.CFG_FILE_NAME)
}
func (*InitCmd) Synopsis() string {
	return fmt.Sprintf(`Initializes a dafault %s at the current directory.`, config.CFG_FILE_NAME)
}

func (ic *InitCmd) Run(args []string) int {
	initcfgfile(config.CfgFileCwd(),ic.Ui)
	return 0
}

func initcfgfile(path string,ui cli.Ui) {
	cfgfile := `
[network]
interface=127.0.0.1

[usertags]
tag=my_tag1:myvalue1
tag=my_tag2:myvalue2
`
	ui.Info(fmt.Sprintf("Initializing config at %s", path))

	f, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {

		if os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(path), 0777)
			if err != nil {
				ui.Error(err.Error())
			}
			f, err = os.Create(path)
		}
		if err != nil {
			ui.Error(err.Error())
		}
	}
	defer f.Close()

	_, err = f.Write([]byte(cfgfile))
	if err != nil {
		ui.Error(err.Error())
	}
	ui.Info("File created sucessfully")
}
