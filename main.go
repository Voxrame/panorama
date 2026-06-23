package main

import (
	"flag"
	"log/slog"
	"os"
	"path"

	"github.com/lord-server/panorama/internal/config"
	"github.com/lord-server/panorama/internal/game"
	"github.com/lord-server/panorama/internal/render"
	"github.com/lord-server/panorama/internal/render/isometric"
	"github.com/lord-server/panorama/internal/tile"
	"github.com/lord-server/panorama/internal/web"
	"github.com/lord-server/panorama/internal/world"
)

type Args struct {
	ConfigPath string
}

var args Args

func fullrender(config config.Config) error {
	descPath := path.Join(config.System.WorldPath, "nodes_dump.json")

	slog.Info("loading game description", "game", config.System.GamePath, "mods", config.System.ModPath, "desc", descPath)

	game, err := game.LoadGame(descPath, config.System.GamePath, config.System.ModPath)
	if err != nil {
		slog.Error("unable to load game description", "error", err)
		return err
	}
	backend, err := world.NewPostgresBackend(config.System.WorldDSN)
	if err != nil {
		slog.Error("unable to connect to world DB", "error", err)
		return err
	}

	world := world.NewWorldWithBackend(backend)

	tiler := tile.NewTiler(config.Region, config.Renderer.ZoomLevels, config.System.TilesPath)

	slog.Info("performing a full render", "workers", config.Renderer.Workers, "region", config.Region)

	tiler.FullRender(&game, &world, config.Renderer.Workers, config.Region, func() render.Renderer {
		return isometric.NewRenderer(config.Region, &game)
	})

	tiler.DownscaleTiles()

	return nil
}

func run(config config.Config) error {
	quit := make(chan bool)

	slog.Info("starting web server", "address", config.Web.ListenAddress)
	go func() {
		web.Serve(&config)
		quit <- true
	}()

	<-quit

	return nil
}

func main() {
	if len(os.Args) < 2 {
		slog.Error("expected a subcommand: (available subcommands: run, fullrender)")
		os.Exit(1)
	}

	subcommand := os.Args[1]

	commonFlags := flag.NewFlagSet("common flags", flag.ExitOnError)
	commonFlags.StringVar(&args.ConfigPath, "config", "config.toml", "Path to config file")
	commonFlags.Parse(os.Args[2:])

	slog.Info("loading config", "config_path", args.ConfigPath)
	config, err := config.LoadConfig(args.ConfigPath)
	if err != nil {
		slog.Error("unable to load config", "error", err)
	}

	switch subcommand {
	case "run":
		err = run(config)

	case "fullrender":
		err = fullrender(config)
	}

	if err != nil {
		os.Exit(1)
	}
}
