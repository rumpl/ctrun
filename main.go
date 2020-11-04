package main

import (
	"fmt"
	"os"

	"github.com/rumpl/ctrun/pkg/server"
	"github.com/rumpl/ctrun/pkg/storage"
	"github.com/rumpl/ctrun/pkg/storage/types"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	var debug bool
	var address string
	var storageOpts types.StorageOpts
	app := &cli.App{
		Name:  "ctrun",
		Usage: "No more docker build",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Usage:       "Debug log level",
				Destination: &debug,
			},
			&cli.StringFlag{
				Name:        "address",
				Aliases:     []string{"a"},
				Usage:       "Address to listen to",
				Value:       "127.0.0.1:1323",
				Destination: &address,
			},
			&cli.StringFlag{
				Name:        "endpoint",
				Usage:       "S3 endpoint",
				EnvVars:     []string{"S3_ENDPOINT"},
				Destination: &storageOpts.Endpoint,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "access-key",
				Usage:       "Access key for the S3 storage",
				EnvVars:     []string{"ACCESS_KEY_ID"},
				Destination: &storageOpts.AccessKey,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "secret-key-id",
				Usage:       "Secret key id for the S3 storage",
				EnvVars:     []string{"SECRET_KEY_ID"},
				Destination: &storageOpts.SecretKey,
				Required:    true,
			}, &cli.StringFlag{
				Name:        "bucket",
				Usage:       "S3 bucket",
				EnvVars:     []string{"S3_BUCKET"},
				Destination: &storageOpts.Bucket,
				Value:       "ctrun",
			},
		},
		Action: func(clix *cli.Context) error {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			store, err := storage.New(clix.Context, storageOpts)
			if err != nil {
				return err
			}

			s := server.New(address, store)

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
