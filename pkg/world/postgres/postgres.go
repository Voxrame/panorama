package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lord-server/panorama/pkg/geom"
	"github.com/lord-server/panorama/pkg/world/selector"
)

type Backend struct {
	conn *pgxpool.Pool
}

func NewBackend(dsn string) (*Backend, error) {
	conn, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	return &Backend{
		conn: conn,
	}, nil
}

func (p *Backend) Close() error {
	p.conn.Close()

	return nil
}

func (p *Backend) GetBlockData(pos geom.BlockPosition) ([]byte, error) {
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

func (p *Backend) GetBlocks(sel selector.BlockSelector, callback func(geom.BlockPosition, []byte) error) error {
	sql, args := sel.Query()

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

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
