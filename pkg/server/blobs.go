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
	logrus.Debugf("Getting blobs for %s", parts[1])

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	blobURL := s.store.Url(ctx, "", parts[1])
	logrus.Infof("Blobs redirecting to %s", blobURL)
	http.Redirect(w, r, blobURL, 301)
}
