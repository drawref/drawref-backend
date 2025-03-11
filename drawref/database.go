package drawref

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DRDatabase struct {
	pool *pgxpool.Pool
}

var TheDb *DRDatabase

func OpenDatabase(connectionUrl string) error {
	pool, err := pgxpool.New(context.Background(), connectionUrl)
	if err != nil {
		return err
	}

	TheDb = &DRDatabase{
		pool,
	}

	return nil
}

func (db *DRDatabase) Close() {
	db.pool.Close()
}
