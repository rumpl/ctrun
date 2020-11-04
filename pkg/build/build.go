package build

import (
	"archive/tar"
	"context"
	"io"
	"net"
	"os"
	"time"

	"github.com/Pallinder/go-randomdata"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/moby/pkg/stdcopy"
	"github.com/pkg/errors"
	"github.com/rumpl/ctrun/pkg/storage/types"
	"golang.org/x/sync/errgroup"
)

const manifestV1 = "application/vnd.oci.image.manifest.v1+json"

const builderImage = "moby/buildkit:buildx-stable-1"

type Client interface {
	Build(string) (string, error)
	Close()
}

type buildClient struct {
	c     *client.Client
	store types.Storage
}

func NewBuilder(store types.Storage) (Client, error) {
	dc, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
	if err != nil {
		return nil, err
	}
	dc.NegotiateAPIVersion(context.Background())
	hc := &container.HostConfig{
		Privileged: true,
	}
	cfg := &container.Config{
		Image: builderImage,
	}
	name := "buildx_buildkit_" + randomdata.SillyName()
	_, err = dc.ContainerCreate(context.Background(), cfg, hc, &network.NetworkingConfig{}, name)
	if err != nil {
		return nil, err
	}
	if err = dc.ContainerStart(context.Background(), name, dockertypes.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	time.Sleep(1 * time.Second)

	// this sucks, this is a buildkit container started by buildx, we need the same "driver" buildx has.
	response, err := dc.ContainerExecCreate(context.Background(), name, dockertypes.ExecConfig{
		Cmd:          []string{"buildctl", "dial-stdio"},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return nil, err
	}
	resp, err := dc.ContainerExecAttach(context.Background(), response.ID, dockertypes.ExecStartCheck{})
	if err != nil {
		return nil, err
	}

	conn := demuxConn(resp.Conn)

	c, err := client.New(context.Background(), "", client.WithDialer(func(string, time.Duration) (net.Conn, error) {
		return conn, nil
	}))
	if err != nil {
		return nil, err
	}

	return &buildClient{
		store: store,
		c:     c,
	}, nil
}

func (b *buildClient) Build(repo string) (string, error) {
	solveOpt := client.SolveOpt{
		Frontend: "dockerfile.v0",
		FrontendAttrs: map[string]string{
			"context":  "git://" + repo,
			"filename": "Dockerfile",
		},
		Exports: []client.ExportEntry{
			{
				Type:   "oci",
				Output: b.wrapWriteCloser(),
			},
		},
	}

	eg, ctx := errgroup.WithContext(context.Background())
	digest := ""

	var def *llb.Definition
	ch := make(chan *client.SolveStatus)
	eg.Go(func() error {
		res, err := b.c.Solve(ctx, def, solveOpt, ch)
		if err != nil {
			return errors.Wrap(err, "solve")
		}

		digest = res.ExporterResponse["containerimage.digest"]
		if digest == "" {
			return errors.New("unable to get the digest of the image")
		}

		return nil
	})

	eg.Go(func() error {
		// Read all the channel so that the build finishes...
		for range ch {
		}
		return nil
	})

	return digest, eg.Wait()
}

func (b *buildClient) Close() {
	b.c.Close()
}

func (b *buildClient) wrapWriteCloser() func(map[string]string) (io.WriteCloser, error) {
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
				if err := b.store.Put(header.Name, tr, manifestV1); err != nil {
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
