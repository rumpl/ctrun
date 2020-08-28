package server

import (
	"net/http"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/gorilla/mux"
)

// ServerOpts are the options for the server
type ServerOpts struct {
	Address     string
	AccessKey   string
	SecretKeyID string
}

type Server interface {
	Start() error
}

type registryBuildServer struct {
	opts ServerOpts
}

func New(opts ServerOpts) Server {
	return &registryBuildServer{
		opts: opts,
	}
}

func (s *registryBuildServer) Start() error {
	router := mux.NewRouter()

	router.HandleFunc("/v2/{name:"+reference.NameRegexp.String()+"}/manifests/{reference}", manifests)
	router.HandleFunc("/v2/{name:"+reference.NameRegexp.String()+"}/blobs/{reference}", blobs)

	srv := &http.Server{
		Handler:      router,
		Addr:         s.opts.Address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return srv.ListenAndServe()
}
