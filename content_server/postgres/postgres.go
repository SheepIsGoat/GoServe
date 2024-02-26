package postgres

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Structs
type PoolConfig struct {
	ConnectionString string
	MaxConns         int32
	MinConns         int32
	ConnectTimeout   time.Duration
}

type PostgresContext struct {
	Pool *pgxpool.Pool
	Ctx  context.Context
}

// Connection pooling
func GetDefaultConfig() *PoolConfig {
	connectionString := os.Getenv("PG_CONN")
	if connectionString == "" {
		connectionString = "postgresql://postgres:mysecretpassword@localhost:5432/server_db"
	}

	maxConnsStr := os.Getenv("MAX_CONNS")
	maxConnsInt, err := strconv.Atoi(maxConnsStr)
	maxConns := int32(maxConnsInt)
	if err != nil {
		maxConns = 10
	}

	timeoutStr := os.Getenv("TIMEOUT")
	timeoutInt, err := strconv.Atoi(timeoutStr)
	timeout := time.Duration(timeoutInt)
	if err != nil {
		timeout = 5
	}

	return &PoolConfig{
		ConnectionString: connectionString,
		MaxConns:         maxConns,
		MinConns:         5,
		ConnectTimeout:   time.Second * timeout,
	}
}

func GetConnectionPool(config *PoolConfig) (*pgxpool.Pool, error) {
	pgxConfig, err := pgxpool.ParseConfig(config.ConnectionString)
	if err != nil {
		return nil, err
	}

	pgxConfig.MaxConns = config.MaxConns
	pgxConfig.MinConns = config.MinConns
	pgxConfig.ConnConfig.ConnectTimeout = config.ConnectTimeout

	pool, err := pgxpool.ConnectConfig(context.Background(), pgxConfig)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
