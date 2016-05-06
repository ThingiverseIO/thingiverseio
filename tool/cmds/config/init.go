package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/ThingiverseIO/thingiverseio/config"
)

var InitCmd = cli.Command{
	Name:  "init",
	Usage: fmt.Sprintf(`Initializes a default %s at the current directory.`, config.CFG_FILE_NAME),

	Description: fmt.Sprintf(`Initializes a default %s at the current directory.

Specify a path or use subcommands to init a %s at different places`, config.CFG_FILE_NAME, config.CFG_FILE_NAME),
	Action: runInitCfgFile,
	Subcommands: []cli.Command{
		InitGlobalCmd,
	},
}
var InitGlobalCmd = cli.Command{
	Name:  "global",
	Usage: fmt.Sprintf(`Initializes a default %s at the global directory.`, config.CFG_FILE_NAME),

	Description: fmt.Sprintf(`Initializes a default %s at %s.`, config.CFG_FILE_NAME, config.CfgFileGlobal()),
	Action:      runInitCfgFileGlobal,
}

func runInitCfgFileGlobal(c *cli.Context) {
	initCfgFile(config.CfgFileGlobal())
}

func runInitCfgFile(c *cli.Context) {
	path := config.CfgFileCwd()
	if c.NArg() > 0 {
		path = c.Args()[0]
	}
	initCfgFile(path)
}

func initCfgFile(path string) {

	log.Printf("Initializing config at %s", path)

	f, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {

		if os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(path), 0777)
			if err != nil {
				log.Println(err.Error())
				return
			}
			f, err = os.Create(path)
		}
		if err != nil {
			log.Println(err.Error())
			return
		}
	}
	defer f.Close()

	_, err = f.Write([]byte(cfgfile))
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("File created sucessfully")
}

var cfgfile = `
[network]
interface=127.0.0.1

[usertags]
tag=my_tag1:myvalue1
tag=my_tag2:myvalue2
`
