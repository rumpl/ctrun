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

	repo := vars["name"]

	logrus.Infof("Manifest %s", repo)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	builder, err := build.NewBuilder(ctx, s.store)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(500)
		return
	}
	defer builder.Close()

	digest, err := builder.Build(ctx, repo)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(500)
		return
	}

	parts := strings.Split(digest, ":")
	blobURL := s.store.URL(ctx, repo, parts[1])
	logrus.Debugf("Image build done, redirecting to %s", blobURL)
	http.Redirect(w, r, blobURL, 301)
}
