package server

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rumpl/ctrun/pkg/build"
	"github.com/sirupsen/logrus"
)

func (s *registryBuildServer) manifests(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	logrus.Info("Manifest ", vars["name"])

	builder, err := build.NewBuilder(s.store)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(500)
		return
	}
	defer builder.Close()

	digest, err := builder.Build(vars["name"])
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(500)
		return
	}

	parts := strings.Split(digest, ":")
	logrus.Info("Done ", s.store.Url(parts[1]))
	http.Redirect(w, r, s.store.Url(parts[1]), 301)
}
