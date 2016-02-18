package main

import (
	"log"
	"os"

	"github.com/joernweissenborn/thingiverse.io/tvio-cfg/cmds"
	"github.com/mitchellh/cli"
)

var logger = log.New(os.Stdout, "", 0)

func main() {
	c := cli.NewCLI("app", "1.0.0")
	c.Args = os.Args[1:]
	ui := &cli.BasicUi{
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}
	c.Commands = map[string]cli.CommandFactory{
		"show": func() (cli.Command, error) {
			return &cmds.ShowCommand{ui}, nil
		},
		"init": func() (cli.Command, error) {
			return &cmds.InitCmd{ui}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
