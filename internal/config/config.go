package config

import (
	"errors"
	"fmt"
	"io"
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

func LoadConfig(path string) (config Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return config, fmt.Errorf("open file: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
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
