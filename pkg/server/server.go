package server

import (
	"net/http"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/gorilla/mux"
	"github.com/rumpl/ctrun/pkg/storage/types"
)

// Opts are the options for the server
type Opts struct {
	Address     string
	AccessKey   string
	SecretKeyID string
}

type Server interface {
	Start() error
}

type registryBuildServer struct {
	address string
	store   types.Storage
}

func New(address string, store types.Storage) Server {
	return &registryBuildServer{
		address: address,
		store:   store,
	}
}

func (s *registryBuildServer) Start() error {
	router := mux.NewRouter()

	router.HandleFunc("/v2/{name:"+reference.NameRegexp.String()+"}/manifests/{reference}", s.manifests)
	router.HandleFunc("/v2/{name:"+reference.NameRegexp.String()+"}/blobs/{reference}", blobs)

	srv := &http.Server{
		Handler:      router,
		Addr:         s.address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return srv.ListenAndServe()
}
