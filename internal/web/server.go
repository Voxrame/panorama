package web

import (
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/lord-server/panorama/frontend"
	"github.com/lord-server/panorama/internal/config"
	"github.com/lord-server/panorama/internal/web/handlers"
)

func Serve(config *config.Config) {
	mux := http.NewServeMux()

	mux.Handle("GET /api/v1/metadata", handlers.Metadata(config))

	if _, err := fs.Stat(frontend.FS(), "index.html"); err == nil {
		mux.Handle("GET /", spaHandler(frontend.FS()))
	} else {
		slog.Warn("frontend dist is empty, SPA routes disabled")
	}

	tilesHandler := withCacheControl(http.StripPrefix("/tiles/", http.FileServer(http.Dir(config.System.TilesPath))))
	mux.Handle("GET /tiles/", tilesHandler)

	srv := &http.Server{
		Addr:    config.Web.ListenAddress,
		Handler: mux,
	}

	if err := srv.ListenAndServe(); err != nil {
		slog.Error("failed to start web server", "err", err)
	}
}

func withCacheControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(5))
		h.ServeHTTP(w, r)
	})
}

func spaHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(fsys))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path != "" {
			if _, err := fs.Stat(fsys, path); err != nil {
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}
