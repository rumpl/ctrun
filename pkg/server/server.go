package server

import (
	"fmt"
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

func Run(opts ServerOpts) error {
	router := mux.NewRouter()

	router.HandleFunc("/v2/{name:"+reference.NameRegexp.String()+"}/manifests/{reference}", manifests)
	router.HandleFunc("/v2/{name:"+reference.NameRegexp.String()+"}/blobs/{reference}", blobs)

	srv := &http.Server{
		Handler:      router,
		Addr:         opts.Address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Server started")
	return srv.ListenAndServe()
}
