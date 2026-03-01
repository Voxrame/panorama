package world

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lord-server/panorama/pkg/geom"
)

type Backend interface {
	GetBlockData(pos geom.BlockPosition) ([]byte, error)
	GetBlocks(selector BlockSelector, callback func(geom.BlockPosition, []byte) error) error
	Close()
}

type PostgresBackend struct {
	conn *pgxpool.Pool
}

func NewPostgresBackend(dsn string) (*PostgresBackend, error) {
	conn, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	return &PostgresBackend{
		conn: conn,
	}, nil
}

func (p *PostgresBackend) Close() {
	p.conn.Close()
}

func (p *PostgresBackend) GetBlockData(pos geom.BlockPosition) ([]byte, error) {
	var data []byte

	err := p.conn.QueryRow(context.Background(), "SELECT data FROM blocks WHERE posx=$1 and posy=$2 and posz=$3", pos.X, pos.Y, pos.Z).Scan(&data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (p *PostgresBackend) GetBlocks(selector BlockSelector, callback func(geom.BlockPosition, []byte) error) error {
	sql, args := selector.Query()

	rows, err := p.conn.Query(context.Background(), sql, args...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			pos  geom.BlockPosition
			data []byte
		)

		err = rows.Scan(&pos.X, &pos.Y, &pos.Z, &data)
		if err != nil {
			return err
		}

		err = callback(pos, data)
		if err != nil {
			return err
		}
	}

	rows.Close()

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
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

func (w *World) GetBlocks(selector BlockSelector, callback func(geom.BlockPosition, *MapBlock) error) error {
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
