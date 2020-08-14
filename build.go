package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/containerd/console"
	dockerclient "github.com/docker/docker/client"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

// build is for testing various buildkit stuff
func build(clix *cli.Context) error {
	eg, ctx := errgroup.WithContext(context.Background())
	dc, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
	if err != nil {
		logrus.Fatal(err)
	}

	c, err := client.New(context.Background(), "", client.WithDialer(func(string, time.Duration) (net.Conn, error) {
		return dc.DialHijack(context.Background(), "/grpc", "h2c", nil)
	}))
	if err != nil {
		logrus.Fatal(err)
	}
	defer c.Close()

	s := llb.Image("alpine").Run(llb.Shlexf("ls -la")).Root()
	// Was a test, doesn't work
	// s := llb.Git("git://github.com/undefinedlabs/hello-world", "master").Run(llb.Shlexf("pwd"))

	def, err := s.Marshal(llb.LinuxAmd64)
	if err != nil {
		logrus.Fatal(err)
	}

	ch := make(chan *client.SolveStatus)
	eg.Go(func() error {
		res, err := c.Solve(ctx, def, client.SolveOpt{
			Exports: []client.ExportEntry{{Type: "moby", Attrs: map[string]string{}}},
		}, ch)
		if err != nil {
			return errors.Wrap(err, "solve")
		}
		for k, v := range res.ExporterResponse {
			fmt.Printf("solve response: %s=%s\n", k, v)
		}
		return nil
	})

	displayCh := ch

	eg.Go(func() error {
		var c console.Console
		return progressui.DisplaySolveStatus(context.TODO(), "", c, os.Stdout, displayCh)
	})

	return eg.Wait()
}
