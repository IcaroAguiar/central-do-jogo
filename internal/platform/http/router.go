package httpplatform

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// Options configures the HTTP router.
type Options struct {
	StaticDir string
	Now       func() time.Time
}

// NewRouter builds the application ServeMux.
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

	if opts.StaticDir != "" {
		fileServer := http.FileServer(http.Dir(opts.StaticDir))
		mux.Handle("GET /", spaHandler(opts.StaticDir, fileServer))
	}

	return mux
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
