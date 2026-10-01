package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed static
var staticFS embed.FS

// ServiceWorkerHandler menyajikan sw.js di root (scope seluruh situs).
func ServiceWorkerHandler() http.HandlerFunc {
	b, err := staticFS.ReadFile("static/sw.js")
	if err != nil {
		panic(err)
	}
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(b)
	}
}

// StaticHandler menyajikan aset statis yang ter-embed (CSS, JS, font, vendor).
func StaticHandler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	fsrv := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/fonts/") || strings.HasPrefix(r.URL.Path, "/static/vendor/") {
			w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		if strings.HasSuffix(r.URL.Path, ".webmanifest") {
			w.Header().Set("Content-Type", "application/manifest+json")
		}
		http.StripPrefix("/static", fsrv).ServeHTTP(w, r)
	})
}
