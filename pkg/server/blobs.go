package server

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func (s *registryBuildServer) blobs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parts := strings.Split(vars["reference"], ":")
	http.Redirect(w, r, s.store.Url(parts[1]), 301)
}
