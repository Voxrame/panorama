package config

import (
	"io"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/lord-server/panorama/pkg/geom"
)

type Web struct {
	ListenAddress string `toml:"listen_address"`
	Title         string `toml:"title"`
}

type Renderer struct {
	Workers    int `toml:"workers"`
	ZoomLevels int `toml:"zoom_levels"`
}

type System struct {
	GamePath  string `toml:"game_path"`
	ModPath   string `toml:"mod_path"`
	TilesPath string `toml:"tiles_path"`
	WorldPath string `toml:"world_path"`
	WorldDSN  string `toml:"world_dsn"`
}

type Config struct {
	System   System      `toml:"system"`
	Web      Web         `toml:"web"`
	Renderer Renderer    `toml:"renderer"`
	Region   geom.Region `toml:"region"`
}

func LoadConfig(path string) (Config, error) {
	var config Config

	file, err := os.Open(path)
	if err != nil {
		return config, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return config, err
	}

	_, err = toml.Decode(string(data), &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
