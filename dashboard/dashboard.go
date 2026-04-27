package dashboard

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/davewhit3/compile-interceptor/outgoing"
)

//go:embed static/index.html
var indexHTML []byte

type Registrar interface {
	Handle(pattern string, handler http.Handler)
}

type RegistrarFunc func(pattern string, handler http.Handler)

func (f RegistrarFunc) Handle(pattern string, handler http.Handler) {
	f(pattern, handler)
}

// Register mounts the Telescope dashboard routes onto mux:
//
//	GET    /telescope                 → serves the browser UI
//	GET    /telescope/api/requests    → returns stored entries as JSON
//	DELETE /telescope/api/requests    → clears the store
func Register(mux Registrar) {
	mux.Handle("/telescope", http.HandlerFunc(handleIndex))
	mux.Handle("/telescope/api/requests", http.HandlerFunc(handleRequests))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(indexHTML)
}

func handleRequests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		entries := outgoing.ListEntries()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(entries)

	case http.MethodDelete:
		outgoing.Reset()
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
