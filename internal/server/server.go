package server

import (
	"errors"
	"io/fs"
	"net/http"
	"time"

	"github.com/Voxrame/panorama/internal/config"
)

func Serve(static fs.FS, config *config.Config) {
	mux := http.NewServeMux()

	staticRootDir, err := fs.Sub(static, "ui/build")
	if err != nil {
		panic(err)
	}

	mux.Handle("/*", http.FileServer(http.FS(staticRootDir)))

	httpServer := &http.Server{
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		Addr:              config.Web.ListenAddress,
		Handler:           mux,
	}

	err = httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
