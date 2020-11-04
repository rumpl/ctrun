package server

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (s *registryBuildServer) blobs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parts := strings.Split(vars["reference"], ":")
	logrus.Info("Blobs ", parts)
	http.Redirect(w, r, s.store.Url(parts[1]), 301)
}
