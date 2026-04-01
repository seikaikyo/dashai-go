package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist/*
var distFS embed.FS

// Handler returns an http.Handler that serves the embedded dashboard SPA.
// All unmatched paths fall back to index.html for client-side routing.
func Handler() http.Handler {
	subFS, _ := fs.Sub(distFS, "dist")
	fileServer := http.FileServer(http.FS(subFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strip /dashboard prefix (chi mounts with prefix)
		path := r.URL.Path

		// Try serving the file directly
		if path != "/" && path != "" {
			cleanPath := strings.TrimPrefix(path, "/")
			if f, err := subFS.Open(cleanPath); err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// SPA fallback: serve index.html
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
