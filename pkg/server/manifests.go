package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rumpl/ctrun/pkg/build"
	"github.com/sirupsen/logrus"
)

func (s *registryBuildServer) manifests(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	logrus.Info("Manifest ", vars["name"])

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	builder, err := build.NewBuilder(ctx, s.store)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(500)
		return
	}
	defer builder.Close()

	digest, err := builder.Build(ctx, vars["name"])
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(500)
		return
	}

	parts := strings.Split(digest, ":")
	logrus.Debugf("Image build done, redirecting to %s", s.store.Url(ctx, parts[1]))
	http.Redirect(w, r, s.store.Url(ctx, parts[1]), 301)
}
