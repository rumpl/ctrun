package main

import (
	"fmt"
	"os"

	"github.com/rumpl/ctrun/pkg/server"
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
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "secret-key-id",
				Usage:       "Secret key id for the S3 storage",
				EnvVars:     []string{"SECRET_KEY_ID"},
				Destination: &opts.SecretKeyID,
				Required:    true,
			},
		},
		Action: func(clix *cli.Context) error {
			s := server.New(opts)

			// It's a lie but we don't care
			fmt.Println("🚀 Server started")

			return s.Start()
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "💀 %s\n", err)
		os.Exit(1)
	}
}
