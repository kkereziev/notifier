package internal

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/uuid"
)

// CORSMiddleware setups CORS for pre-flight requests.
func CORSMiddleware(next httptreemux.HandlerFunc) httptreemux.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, m map[string]string) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)

			return
		}

		next(w, r, m)
	}
}

// RecoverMiddleware recovers the server from unexpected crashes.
func RecoverMiddleware(logger *log.Logger) func(httptreemux.HandlerFunc) httptreemux.HandlerFunc {
	return func(next httptreemux.HandlerFunc) httptreemux.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, m map[string]string) {
			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {
					trace := debug.Stack()

					// Stack trace will be provided.
					logger.Printf("PANIC [%v] TRACE[%s]", rec, trace)
				}
			}()

			next(w, r, m)
		}
	}
}

// LoggingMiddleware logs request and response objects.
func LoggingMiddleware(logger *log.Logger) func(httptreemux.HandlerFunc) httptreemux.HandlerFunc {
	return func(next httptreemux.HandlerFunc) httptreemux.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, m map[string]string) {
			traceID := uuid.New()
			tNow := time.Now().UTC()

			log.Println(
				"request started", "trace id", traceID, "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr,
			)

			next(w, r, m)

			log.Println("request completed", "trace ID", traceID, "method", r.Method,
				"path", r.URL.Path, "remoteaddr", r.RemoteAddr, "since", time.Since(tNow))
		}
	}
}
