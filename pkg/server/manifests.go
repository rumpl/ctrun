package server

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/moby/pkg/stdcopy"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func manifests(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	dc, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
	if err != nil {
		panic(err)
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
		panic(err)
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
	digest := ""
	var def *llb.Definition
	eg.Go(func() error {
		res, err := c.Solve(ctx, def, solveOpt, ch)
		if err != nil {
			return errors.Wrap(err, "solve")
		}

		digest = res.ExporterResponse["containerimage.digest"]

		return nil
	})

	displayCh := ch

	eg.Go(func() error {
		for range displayCh {
		}
		return nil
	})

	if err = eg.Wait(); err != nil {
		panic(err)
	}

	ss := strings.Split(digest, ":")
	dd := ss[1]
	http.Redirect(w, r, fmt.Sprintf("https://ctrun.s3.fr-par.scw.cloud/blobs/sha256/%s", dd), 301)
}

func wrapWriteCloser() func(map[string]string) (io.WriteCloser, error) {
	endpoint := "s3.fr-par.scw.cloud"
	accessKeyID := os.Getenv("ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("SECRET_KEY_ID")
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		panic(err)
	}

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
				panic(err)
			case header == nil:
				return
			}

			switch header.Typeflag {
			case tar.TypeDir:
				continue
			case tar.TypeReg:
				bucket := "ctrun"
				_, err = minioClient.PutObject(context.Background(), bucket, header.Name, tr, -1, minio.PutObjectOptions{
					UserMetadata: map[string]string{
						"x-amz-acl": "public-read",
					},
					ContentType: "application/vnd.oci.image.manifest.v1+json",
				})
				if err != nil {
					panic(err)
				}
			}
		}
	}()

	return func(d map[string]string) (io.WriteCloser, error) {
		return pw, nil
	}
}

func demuxConn(c net.Conn) net.Conn {
	pr, pw := io.Pipe()
	// nolint: errcheck
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
