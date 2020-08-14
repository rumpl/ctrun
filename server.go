package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/docker/distribution/reference"
	dockerclient "github.com/docker/docker/client"
	"github.com/go-git/go-git/v5"
	"github.com/gorilla/mux"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

// server is the most awesomest registry
func server(clix *cli.Context) error {
	router := mux.NewRouter()

	router.HandleFunc("/v2/{name:"+reference.NameRegexp.String()+"}/manifests/{reference}", manifests)

	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:1323",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return srv.ListenAndServe()
}

func manifests(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	_, _ = git.PlainClone("/tmp/foo", false, &git.CloneOptions{
		URL:      "https://" + vars["name"],
		Progress: os.Stdout,
	})

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
	ch := make(chan *client.SolveStatus)

	solveOpt := client.SolveOpt{
		Frontend: "dockerfile.v0",
		Exports:  []client.ExportEntry{{Type: "moby", Attrs: map[string]string{}}},
	}
	solveOpt.LocalDirs, err = attrMap(
		fmt.Sprintf("context=%s", "/tmp/foo"),
		fmt.Sprintf("dockerfile=%s", "/tmp/foo"),
	)
	if err != nil {
		logrus.Fatal("attrmap", err)
	}

	eg, ctx := errgroup.WithContext(context.Background())

	var def *llb.Definition
	eg.Go(func() error {
		res, err := c.Solve(ctx, def, solveOpt, ch)
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
		for range displayCh {
		}
		return nil
	})

	if err = eg.Wait(); err != nil {
		logrus.Fatal(err)
	}
}

func attrMap(sl ...string) (map[string]string, error) {
	m := map[string]string{}
	for _, v := range sl {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			return nil, errors.Errorf("invalid value %s", v)
		}
		m[parts[0]] = parts[1]
	}
	return m, nil
}
