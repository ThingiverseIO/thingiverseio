package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "tvio"
	app.Usage = "The swiss army knife for thingiverse.io"
//	app.Action = func(c *cli.Context) {
//		println("boom! I say!")
//	}

	app.Run(os.Args)
}
