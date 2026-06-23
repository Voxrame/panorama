package web

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/lord-server/panorama/internal/config"
	"github.com/lord-server/panorama/internal/web/handlers"
)

func Serve(config *config.Config) {
	mux := http.NewServeMux()

	mux.Handle("GET /api/v1/metadata", handlers.Metadata(config))

	mux.Handle("GET /", http.FileServer(http.Dir("./static")))

	tilesHandler := withCacheControl(http.StripPrefix("/tiles/", http.FileServer(http.Dir(config.System.TilesPath))))
	mux.Handle("GET /tiles/", tilesHandler)

	srv := &http.Server{
		Addr:    config.Web.ListenAddress,
		Handler: mux,
	}

	slog.Info("starting web server", "address", config.Web.ListenAddress)
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
