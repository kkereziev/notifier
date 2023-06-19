package internal_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kkereziev/notifier/internal"
)

func TestMux(t *testing.T) {
	server := internal.NewMux()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/te", nil)
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(b))
	t.Log(req)
}
