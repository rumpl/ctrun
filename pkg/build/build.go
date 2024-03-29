package build

import (
	"archive/tar"
	"context"
	"fmt"
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
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const manifestV1 = "application/vnd.oci.image.manifest.v1+json"

const builderImage = "moby/buildkit:master"

type Client interface {
	Build(context.Context, string) (string, error)
	Close()
}

type buildClient struct {
	c     *client.Client
	store types.Storage
}

func NewBuilder(ctx context.Context, store types.Storage) (Client, error) {
	dc, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
	if err != nil {
		return nil, err
	}

	dc.NegotiateAPIVersion(ctx)

	hc := &container.HostConfig{
		Privileged: true,
	}
	cfg := &container.Config{
		Image: builderImage,
	}

	name := "buildx_buildkit_" + randomdata.SillyName()
	if _, err = dc.ContainerCreate(ctx, cfg, hc, &network.NetworkingConfig{}, name); err != nil {
		return nil, err
	}

	if err = dc.ContainerStart(ctx, name, dockertypes.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	// TODO: wait for the daemon to really start, this works, for now
	time.Sleep(1 * time.Second)

	response, err := dc.ContainerExecCreate(ctx, name, dockertypes.ExecConfig{
		Cmd:          []string{"buildctl", "dial-stdio"},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return nil, err
	}

	resp, err := dc.ContainerExecAttach(ctx, response.ID, dockertypes.ExecStartCheck{})
	if err != nil {
		return nil, err
	}

	conn := demuxConn(resp.Conn)

	c, err := client.New(ctx, "", client.WithDialer(func(string, time.Duration) (net.Conn, error) {
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

func (b *buildClient) Build(ctx context.Context, repo string) (string, error) {
	reader, wrapped := b.wrapWriteCloser(ctx, repo)
	solveOpt := client.SolveOpt{
		Frontend: "dockerfile.v0",
		FrontendAttrs: map[string]string{
			"context":  "git://" + repo,
			"filename": "Dockerfile",
		},
		Exports: []client.ExportEntry{
			{
				Type:   "oci",
				Output: wrapped,
			},
		},
		// TODO: This uses an insecure registry as cache, we need to send the right
		// configuration to buildkitd when we create a new instance.
		// CacheExports: []client.CacheOptionsEntry{
		// 	{
		// 		Type: "registry",
		// 		Attrs: map[string]string{
		// 			"ref": "host.docker.internal:5000/cache:1",
		// 		},
		// 	},
		// },
		// CacheImports: []client.CacheOptionsEntry{
		// 	{
		// 		Type: "registry",
		// 		Attrs: map[string]string{
		// 			"ref":  "host.docker.internal:5000/cache:1",
		// 			"mode": "max",
		// 		},
		// 	},
		// },
	}

	eg, _ := errgroup.WithContext(ctx)
	digest := ""

	var def *llb.Definition
	ch := make(chan *client.SolveStatus)

	eg.Go(reader)

	eg.Go(func() error {
		res, err := b.c.Solve(ctx, def, solveOpt, ch)
		logrus.Info("Build finished")
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

func (b *buildClient) wrapWriteCloser(ctx context.Context, repo string) (func() error, func(map[string]string) (io.WriteCloser, error)) {
	pr, pw := io.Pipe()
	reader := func() error {
		tr := tar.NewReader(pr)

		for {
			header, err := tr.Next()
			switch {
			case err == io.EOF:
				return nil
			case err != nil:
				logrus.Errorf("Reading tar contents failed %s", err)
				return err
			case header == nil:
				return nil
			}

			switch header.Typeflag {
			case tar.TypeDir:
				continue
			case tar.TypeReg:
				name := fmt.Sprintf("%s/%s", repo, header.Name)
				if err := b.store.Put(ctx, name, tr, manifestV1); err != nil {
					logrus.Errorf("Put to storage failed %s", err)
					return err
				}
			}
		}
	}

	return reader, func(d map[string]string) (io.WriteCloser, error) {
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
