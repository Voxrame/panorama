package world

import (
	"errors"
	"fmt"
	"path/filepath"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/lord-server/panorama/pkg/geom"
	"github.com/lord-server/panorama/pkg/world/postgres"
	"github.com/lord-server/panorama/pkg/world/selector"
	"github.com/lord-server/panorama/pkg/world/sqlite3"
)

type Backend interface {
	GetBlockData(pos geom.BlockPosition) ([]byte, error)
	GetBlocks(selector selector.BlockSelector, callback func(geom.BlockPosition, []byte) error) error
	Close() error
}

type PostgresBackend = postgres.Backend

func NewPostgresBackend(dsn string) (*PostgresBackend, error) {
	return postgres.NewBackend(dsn)
}

type World struct {
	backend           Backend
	decodedBlockCache *lru.Cache[geom.BlockPosition, *MapBlock]
}

func NewWorld(path string) (World, error) {
	var world World

	meta, err := ParseMeta(filepath.Join(path, "world.mt"))
	if err != nil {
		return world, err
	}

	backendName, ok := meta["backend"]
	if !ok {
		return world, errors.New("backend not specified")
	}

	var backend Backend

	switch backendName {
	case "postgresql":
		dsn, ok := meta["pgsql_connection"]
		if !ok {
			return world, errors.New("postgresql connection not specified")
		}

		backend, err = NewPostgresBackend(dsn)
		if err != nil {
			return world, fmt.Errorf("unable to create PostgreSQL backend: %w", err)
		}

	case "sqlite3":
		dsn := filepath.Join(path, "map.sqlite")

		backend, err = sqlite3.NewBackend(dsn)
		if err != nil {
			return world, fmt.Errorf("unable to create PostgreSQL backend: %w", err)
		}
	}

	return NewWorldWithBackend(backend), nil
}

func NewWorldWithBackend(backend Backend) World {
	decodedBlockCache, err := lru.New[geom.BlockPosition, *MapBlock](1024 * 16)
	if err != nil {
		panic(err)
	}

	return World{
		backend:           backend,
		decodedBlockCache: decodedBlockCache,
	}
}

func (w *World) GetBlock(pos geom.BlockPosition) (*MapBlock, error) {
	cachedBlock, ok := w.decodedBlockCache.Get(pos)

	if ok {
		if cachedBlock == nil {
			return nil, nil
		}

		return cachedBlock, nil
	}

	data, err := w.backend.GetBlockData(pos)
	if err != nil {
		return nil, err
	}

	if data == nil {
		w.decodedBlockCache.Add(pos, nil)
		return nil, nil
	}

	block, err := DecodeMapBlock(data)
	if err != nil {
		return nil, err
	}

	w.decodedBlockCache.Add(pos, block)

	return block, nil
}

func (w *World) GetBlocks(selector selector.BlockSelector, callback func(geom.BlockPosition, *MapBlock) error) error {
	return w.backend.GetBlocks(selector, func(pos geom.BlockPosition, data []byte) error {
		cachedBlock, ok := w.decodedBlockCache.Get(pos)

		if ok {
			if cachedBlock == nil {
				return nil
			}

			return callback(pos, cachedBlock)
		}

		block, err := DecodeMapBlock(data)
		if err != nil {
			return err
		}

		w.decodedBlockCache.Add(pos, block)

		return callback(pos, block)
	})
}
