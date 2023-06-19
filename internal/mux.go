package internal

import (
	"fmt"
	"net/http"

	"github.com/dimfeld/httptreemux/v5"
)

const (
	_apiURI = "/api/v1"
)

// NewMux is a constructor function for creating new multiplexer for the HTTP server.
func NewMux() *httptreemux.TreeMux {
	m := httptreemux.New()

	g := m.NewGroup(_apiURI)

	g.POST("/slack", func(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, "Hi")
	})

	return m
}
