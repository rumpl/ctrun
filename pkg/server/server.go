package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/gorilla/mux"
)

func Run() error {
	router := mux.NewRouter()

	router.HandleFunc("/v2/{name:"+reference.NameRegexp.String()+"}/manifests/{reference}", manifests)
	router.HandleFunc("/v2/{name:"+reference.NameRegexp.String()+"}/blobs/{reference}", blobs)

	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:1323",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Server started")
	return srv.ListenAndServe()
}
