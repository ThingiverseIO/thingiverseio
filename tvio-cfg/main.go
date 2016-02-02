package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/joernweissenborn/thingiverse.io/tvio-cfg/cmds"
	"github.com/mitchellh/cli"
)

var logger = log.New(os.Stdout, "", 0)

func main() {
	c := cli.NewCLI("app", "1.0.0")
	c.Args = os.Args[1:]
	ui := &cli.BasicUi{
		Writer: os.Stdout,
		ErrorWriter: os.Stderr,
	}
	c.Commands = map[string]cli.CommandFactory{
		"show": cmds.ShowCommandFactory,
		"init": func() (cli.Command, error){
			return &cmds.InitCmd{ui}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
func mdain() {
	flag.Parse()
	switch flag.Arg(0) {
	case "show":
		show()
	case "init":
		initcfg()
	default:
		logger.Println("Unknown command, usage: tviocfg [show]")
	}
}

func show() {
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
}

func initcfg() {
	var dir string

	switch flag.Arg(1) {
	case "global":
		dir = config.CfgFileGlobal()
	case "user":
		dir = config.CfgFileUser()
	default:
		dir = config.CfgFileCwd()
	}

	initcfgfile(dir)
}

func initcfgfile(path string) {
	cfgfile := `
[network]
interface=127.0.0.1

[usertags]
tag=my_tag1:myvalue1
tag=my_tag2:myvalue2
`
	logger.Println("Initializing config at", path)

	f, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {

		if os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(path), 0777)
			if err != nil {
				logger.Fatal(err)
			}
			f, err = os.Create(path)
		}
		if err != nil {
			logger.Fatal(err)
		}
	}
	defer f.Close()

	_, err = f.Write([]byte(cfgfile))
	if err != nil {
		logger.Fatal(err)

	}
	logger.Println("Done")
}
