package main

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/moby/pkg/stdcopy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

// server is the most awesomest registry
func server(clix *cli.Context) error {
	router := mux.NewRouter()

	router.HandleFunc("/v2/{name:"+reference.NameRegexp.String()+"}/manifests/{reference}", manifests)
	router.HandleFunc("/v2/{name:"+reference.NameRegexp.String()+"}/blobs/{reference}", manifests)

	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:1323",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return srv.ListenAndServe()
}

func demuxConn(c net.Conn) net.Conn {
	pr, pw := io.Pipe()
	go stdcopy.StdCopy(pw, os.Stderr, c)
	return &demux{
		Conn:   c,
		Reader: pr,
	}
}

type demux struct {
	net.Conn
	io.Reader
}

func (d *demux) Read(dt []byte) (int, error) {
	return d.Reader.Read(dt)
}

func manifests(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	dc, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
	if err != nil {
		logrus.Fatal(err)
	}
	dc.NegotiateAPIVersion(context.Background())
	response, err := dc.ContainerExecCreate(context.Background(), "buildx_buildkit_objective_noyce0", types.ExecConfig{
		Cmd:          []string{"buildctl", "dial-stdio"},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		panic(err)
	}
	resp, err := dc.ContainerExecAttach(context.Background(), response.ID, types.ExecStartCheck{})
	if err != nil {
		panic(err)
	}

	conn := demuxConn(resp.Conn)

	c, err := client.New(context.Background(), "", client.WithDialer(func(string, time.Duration) (net.Conn, error) {
		return conn, nil
	}))
	if err != nil {
		logrus.Fatal(err)
	}
	defer c.Close()
	ch := make(chan *client.SolveStatus)

	solveOpt := client.SolveOpt{
		Frontend: "dockerfile.v0",
		FrontendAttrs: map[string]string{
			"context":  "git://" + vars["name"],
			"filename": "Dockerfile",
		},
		Exports: []client.ExportEntry{
			{
				Type:   "oci",
				Output: wrapWriteCloser(),
			},
		},
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

func wrapWriteCloser() func(map[string]string) (io.WriteCloser, error) {
	pr, pw := io.Pipe()
	// TODO: check errors
	go func() {
		tr := tar.NewReader(pr)

		for {
			header, err := tr.Next()
			switch {
			case err == io.EOF:
				return
			case err != nil:
				return
			case header == nil:
				return
			}

			target := filepath.Join("/tmp/that", header.Name)
			switch header.Typeflag {
			case tar.TypeDir:
				if _, err := os.Stat(target); err != nil {
					if err := os.MkdirAll(target, 0755); err != nil {
						return
					}
				}
			case tar.TypeReg:
				f, err := os.OpenFile(target, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
				if err != nil {
					return
				}
				if _, err := io.Copy(f, tr); err != nil {
					return
				}
				f.Close()
			}
		}
	}()

	return func(d map[string]string) (io.WriteCloser, error) {
		return pw, nil
	}
}
