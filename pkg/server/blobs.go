package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func blobs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	s := strings.Split(vars["reference"], ":")
	dd := s[1]

	http.Redirect(w, r, fmt.Sprintf("https://ctrun.s3.fr-par.scw.cloud/blobs/sha256/%s", dd), 301)
}
