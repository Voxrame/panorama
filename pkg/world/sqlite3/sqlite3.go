package sqlite3

import (
	"database/sql"
	"errors"
	"log"

	"github.com/lord-server/panorama/pkg/geom"
	"github.com/lord-server/panorama/pkg/world/selector"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type Backend struct {
	db *sql.DB
}

func NewBackend(dsn string) (*Backend, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Printf("sqlite3: failed to ping database: %v", err)
		_ = db.Close()
		return nil, err
	}

	return &Backend{db: db}, nil
}

func (b *Backend) Close() error {
	return b.db.Close()
}

func (b *Backend) GetBlockData(pos geom.BlockPosition) ([]byte, error) {
	var data []byte

	err := b.db.QueryRow("SELECT data FROM blocks WHERE posx=$1 AND posy=$2 AND posz=$3", pos.X, pos.Y, pos.Z).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (b *Backend) GetBlocks(selector selector.BlockSelector, callback func(geom.BlockPosition, []byte) error) error {
	query, args := selector.Query()

	rows, err := b.db.Query(query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}

	if err != nil {
		return err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("sqlite3: failed to close rows: %v", err)
		}
	}()

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

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
