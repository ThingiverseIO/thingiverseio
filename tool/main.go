package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/ThingiverseIO/thingiverseio/tool/cmds/config"
)

func main() {
	app := cli.NewApp()
	app.Name = "tvio"
	app.Usage = "The swiss army knife for thingiverse.io"
	app.Commands = []cli.Command{
		config.Command,
	}
	//app.Version = thingiverseio.CurrentVersion.String()
	//	app.Action = func(c *cli.Context) {
	//		println("boom! I say!")
	//	}

	app.Run(os.Args)
}
