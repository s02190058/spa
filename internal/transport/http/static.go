package http

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

type staticHandlers struct {
	path  string
	index string
}

func registerStaticHandlers(r *mux.Router, path, index string) {
	h := staticHandlers{
		path:  path,
		index: index,
	}

	r.PathPrefix("/").HandlerFunc(h.serveFile)
}

func (h *staticHandlers) serveFile(w http.ResponseWriter, r *http.Request) {
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}

	path = filepath.Join(h.path, path)
	_, err = os.Stat(path)
	switch {
	case err == nil:
		http.ServeFile(w, r, path)
	case os.IsNotExist(err):
		http.ServeFile(w, r, filepath.Join(h.path, h.index))
	default:
		errorResponse(w, http.StatusInternalServerError, err)
	}
}
