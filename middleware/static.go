package middleware

import (
	"net/http"
	"path/filepath"
	"strings"
)

// staticExtensions is the allowlist of file types served to browsers. The
// templates and assets directories also hold Go sources and server-rendered
// templates, which must not be downloadable.
var staticExtensions = map[string]bool{
	".css":  true,
	".js":   true,
	".map":  true,
	".png":  true,
	".webp": true,
	".ico":  true,
	".svg":  true,
}

// StaticFiles wraps a file server so only allowlisted extensions are served.
// Everything else, including Go sources and HTML templates, returns 404.
func StaticFiles(fs http.FileSystem) http.Handler {
	fileServer := http.FileServer(fs)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}
		if !staticExtensions[strings.ToLower(filepath.Ext(r.URL.Path))] {
			http.NotFound(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}
