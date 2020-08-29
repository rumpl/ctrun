package server

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rumpl/ctrun/pkg/build"
)

func (s *registryBuildServer) manifests(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	builder, err := build.NewBuilder(s.store)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	defer builder.Close()

	digest, err := builder.Build(vars["name"])
	if err != nil {
		w.WriteHeader(500)
		return
	}

	parts := strings.Split(digest, ":")
	http.Redirect(w, r, s.store.Url(parts[1]), 301)
}
