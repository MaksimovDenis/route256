package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pools struct {
	Master  *pgxpool.Pool
	Replica *pgxpool.Pool
}

func New(ctx context.Context, masterDSN, replicaDSN string) (*Pools, error) {
	masterPool, err := createPool(ctx, masterDSN, "master")
	if err != nil {
		return nil, err
	}

	replicaPool, err := createPool(ctx, replicaDSN, "replica")
	if err != nil {
		return nil, err
	}

	return &Pools{
		Master:  masterPool,
		Replica: replicaPool,
	}, nil
}

func (p *Pools) GetWriteReplica() *pgxpool.Pool {
	return p.Master
}

func (p *Pools) GetReadReplica() *pgxpool.Pool {
	return p.Replica
}

func createPool(ctx context.Context, dsn string, role string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig (%s): %w", role, err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig failed to create %s pool: %w", role, err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%s pool.Ping failed: %w", role, err)
	}

	return pool, nil
}
