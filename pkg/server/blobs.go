package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (s *registryBuildServer) blobs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parts := strings.Split(vars["reference"], ":")
	logrus.Infof("Getting blobs for %s", parts)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	http.Redirect(w, r, s.store.Url(ctx, parts[1]), 301)
}
