package httpplatform

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// SSRHandlers wires optional server-side HTML routes (PAT-004).
// Any handler left nil falls through to the SPA shell.
type SSRHandlers struct {
	Home  http.Handler
	Club  http.Handler
	Match http.Handler
}

// Options configures the HTTP router.
type Options struct {
	StaticDir string
	Now       func() time.Time
	SSR       SSRHandlers
	// RegisterAPI mounts feature API routes onto the mux. The composition
	// root (internal/app) supplies this so the HTTP kernel does not grow a
	// field-per-handler dependency bag (ADR 0002).
	RegisterAPI func(mux *http.ServeMux)
}

// NewRouter builds the application ServeMux. API routes are registered as
// exact/one-segment patterns, which net/http's ServeMux always prefers over
// the "/" catch-all used for the SPA shell, so API routes take priority.
func NewRouter(opts Options) http.Handler {
	if opts.Now == nil {
		opts.Now = time.Now
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":    "ok",
			"checkedAt": opts.Now().UTC().Format(time.RFC3339),
		})
	})

	if opts.RegisterAPI != nil {
		opts.RegisterAPI(mux)
	}
	registerSSRRoutes(mux, opts.SSR)

	if opts.StaticDir != "" {
		fileServer := http.FileServer(http.Dir(opts.StaticDir))
		mux.Handle("GET /", spaHandler(opts.StaticDir, fileServer))
	}

	return mux
}

func registerSSRRoutes(mux *http.ServeMux, ssr SSRHandlers) {
	if ssr.Home != nil {
		// "/{$}" matches only the exact root path, letting the SPA "/"
		// catch-all continue to serve every other client-side route.
		mux.Handle("GET /{$}", ssr.Home)
	}
	if ssr.Club != nil {
		mux.Handle("GET /clubes/{slug}", ssr.Club)
	}
	if ssr.Match != nil {
		mux.Handle("GET /jogos/{slug}", ssr.Match)
	}
}

func spaHandler(staticDir string, fileServer http.Handler) http.Handler {
	root := http.Dir(staticDir)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && !strings.Contains(filepath.Base(r.URL.Path), ".") {
			rel := strings.TrimPrefix(filepath.Clean("/"+r.URL.Path), "/")
			if rel != "" && rel != "." {
				f, err := root.Open(rel)
				if err != nil {
					http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
					return
				}
				_ = f.Close()
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}
