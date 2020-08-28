package main

import (
	"os"

	"github.com/rumpl/ctrun/pkg/server"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "ctrun",
		Action: func(clix *cli.Context) error {
			return server.Run()
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
