package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "ctrun",
		Commands: []*cli.Command{
			{
				Name:   "server",
				Action: server,
			},
			{
				Name:   "build",
				Action: build,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
