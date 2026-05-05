package main

import (
	"log/slog"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/Voxrame/panorama/internal/config"
	"github.com/Voxrame/panorama/internal/server"
	"github.com/Voxrame/panorama/static"
)

type RunArgs struct{}

var args struct {
	ConfigPath string   `arg:"-c,--config" default:"config.toml"`
	Run        *RunArgs `arg:"subcommand:run"`
}

func main() {
	arg.MustParse(&args)

	config, err := config.LoadConfig(args.ConfigPath)
	if err != nil {
		slog.Error("unable to load config", "error", err)
		os.Exit(1)
	}

	switch {
	case args.Run != nil:
		err = run(config)

	default:
		slog.Warn("command not specified, proceeding with run")

		err = run(config)
	}

	if err != nil {
		os.Exit(1)
	}
}
func run(config config.Config) error {
	quit := make(chan bool)

	slog.Info("starting web server", "address", config.Web.ListenAddress)

	go func() {
		server.Serve(static.UI, &config)
		quit <- true
	}()

	<-quit

	return nil
}
