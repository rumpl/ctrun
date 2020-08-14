module github.com/rumpl/ctrun

go 1.14

require (
	github.com/containerd/console v0.0.0-20191219165238-8375c3424e4d
	github.com/docker/buildx v0.4.1
	github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c
	github.com/go-git/go-git v4.7.0+incompatible
	github.com/go-git/go-git/v5 v5.1.0
	github.com/gorilla/mux v1.7.2
	github.com/moby/buildkit v0.7.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/cli v1.22.2
	github.com/urfave/cli/v2 v2.2.0
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.0-beta.2.0.20200728183644-eb6354a11860
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200310163718-4634ce647cf2+incompatible
	github.com/hashicorp/go-immutable-radix => github.com/tonistiigi/go-immutable-radix v0.0.0-20170803185627-826af9ccf0fe
	github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305
)
