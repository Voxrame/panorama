package world

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Meta map[string]string

func ParseMeta(path string) (meta Meta, err error) {
	meta = make(map[string]string)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("can't read world metadata: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	sc := bufio.NewScanner(file)

	for sc.Scan() {
		parts := strings.SplitN(sc.Text(), "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid key-value pair: %v", sc.Text())
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("file contains empty key")
		}

		meta[key] = value
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}

	return meta, nil
}
