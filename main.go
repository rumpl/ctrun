package main

import (
	"os"

	"github.com/rumpl/ctrun/pkg/server"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	var opts server.ServerOpts
	app := &cli.App{
		Name:  "ctrun",
		Usage: "No more docker build",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "address",
				Aliases:     []string{"a"},
				Usage:       "Address to listen to",
				Value:       "127.0.0.1:1323",
				Destination: &opts.Address,
			},
			&cli.StringFlag{
				Name:        "access-key",
				Usage:       "Access key for the S3 storage",
				EnvVars:     []string{"ACCESS_KEY_ID"},
				Destination: &opts.AccessKey,
			},
			&cli.StringFlag{
				Name:        "secret-key-id",
				Usage:       "Secret key id for the S3 storage",
				EnvVars:     []string{"SECRET_KEY_ID"},
				Destination: &opts.SecretKeyID,
			},
		},
		Action: func(clix *cli.Context) error {
			return server.Run(opts)
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
