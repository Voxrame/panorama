package config

import (
	"fmt"
	"io"
	"log"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Web Web `toml:"web"`
}

type Web struct {
	ListenAddress string `toml:"listen_address"`
	Title         string `toml:"title"`
}

func LoadConfig(path string) (Config, error) {
	var config Config

	file, err := os.Open(path)
	if err != nil {
		return config, fmt.Errorf("open file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return config, fmt.Errorf("read file: %w", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("unmarshal yaml: %w", err)
	}

	return config, nil
}
