package httpapi

import (
	"io"
	"net/http"

	"github.com/heidericklucas/go-security-swe-bench/app/safepath"
)

func (s *server) getFile(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	rc, err := s.openFile(name)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer rc.Close()
	_, _ = io.Copy(w, rc)
}

func (s *server) openFile(name string) (io.ReadCloser, error) {
	return safepath.Open(s.filesDir, name)
}
