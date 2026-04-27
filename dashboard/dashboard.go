package dashboard

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/davewhit3/compile-interceptor/outgoing"
)

//go:embed templates static
var assets embed.FS

// pageData is passed to the HTML template on every request.
// It lets the server inject API URLs and asset paths into the page,
// so the dashboard works correctly regardless of where it is mounted.
type pageData struct {
	AssetBase string // base path for static assets, e.g. "/telescope"
	HTTPAPI   string // full URL for the HTTP entries endpoint
	CacheAPI  string // full URL for the cache entries endpoint
}

var (
	indexTmpl  *template.Template
	cssContent []byte
	jsContent  []byte
)

func init() {
	indexTmpl = template.Must(
		template.ParseFS(assets, "templates/index.html"),
	)

	var err error
	cssContent, err = fs.ReadFile(assets, "static/style.css")
	if err != nil {
		panic("telescope: missing static/style.css: " + err.Error())
	}
	jsContent, err = fs.ReadFile(assets, "static/app.js")
	if err != nil {
		panic("telescope: missing static/app.js: " + err.Error())
	}
}

// RouteRegistrar abstracts any HTTP router so that Telescope can be mounted on
// *http.ServeMux, gin, echo, chi, or any other framework without importing it.
//
// Use ForMux for the standard library. For other routers implement the
// interface directly, or pass a function literal:
//
//	// gin:
//	dashboard.Register(func(method, path string, h http.HandlerFunc) {
//	    router.Handle(method, path, gin.WrapH(h))
//	})
//
//	// echo:
//	dashboard.Register(func(method, path string, h http.HandlerFunc) {
//	    e.Add(method, path, echo.WrapHandler(h))
//	})
//
//	// chi:
//	dashboard.Register(func(method, path string, h http.HandlerFunc) {
//	    r.MethodFunc(method, path, h)
//	})
type RouteRegistrar func(method, path string, handler http.HandlerFunc)

// ForMux adapts *http.ServeMux using Go 1.22+ method+path routing patterns.
func ForMux(mux *http.ServeMux) RouteRegistrar {
	return func(method, path string, h http.HandlerFunc) {
		mux.HandleFunc(method+" "+path, h)
	}
}

// Register mounts the Telescope dashboard routes using reg.
func Register(reg RouteRegistrar) {
	reg("GET", "/telescope",                 handleIndex)
	reg("GET", "/telescope/style.css",       handleCSS)
	reg("GET", "/telescope/app.js",          handleJS)
	reg("GET", "/telescope/api/requests",    serveRequests)
	reg("DELETE", "/telescope/api/requests", clearRequests)
	reg("GET", "/telescope/api/cache",       serveCache)
	reg("DELETE", "/telescope/api/cache",    clearCache)
}

func handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = indexTmpl.Execute(w, pageData{
		AssetBase: "/telescope",
		HTTPAPI:   "/telescope/api/requests",
		CacheAPI:  "/telescope/api/cache",
	})
}

func handleCSS(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write(cssContent)
}

func handleJS(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	_, _ = w.Write(jsContent)
}

func serveRequests(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(outgoing.ListRequests())
}

func clearRequests(w http.ResponseWriter, _ *http.Request) {
	outgoing.ResetRequests()
	w.WriteHeader(http.StatusNoContent)
}

func serveCache(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(outgoing.ListCommands())
}

func clearCache(w http.ResponseWriter, _ *http.Request) {
	outgoing.ResetCommands()
	w.WriteHeader(http.StatusNoContent)
}
