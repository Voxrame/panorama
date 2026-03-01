package server

import (
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/lord-server/panorama/internal/config"
)

func Serve(static fs.FS, config *config.Config) {
	mux := http.NewServeMux()

	staticRootDir, err := fs.Sub(static, "ui/build")
	if err != nil {
		panic(err)
	}

	mux.Handle("/*", http.FileServer(http.FS(staticRootDir)))
	mux.Handle("/tiles/*", http.StripPrefix("/tiles", http.FileServer(http.Dir(config.System.TilesPath))))

	httpServer := &http.Server{
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		Addr:              config.Web.ListenAddress,
		Handler:           mux,
	}

	err = httpServer.ListenAndServe()
	if err != nil {
		slog.Error("failed to start web server", "err", err)
	}
}
