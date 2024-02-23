package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
)

// Structs
type PoolConfig struct {
	ConnectionString string
	MaxConns         int32
	MinConns         int32
	ConnectTimeout   time.Duration
}

type UserAuth struct {
	Email           string
	Password        string
	ConfirmPassword string
	Consent         string
}

type PostgresContext struct {
	pool *pgxpool.Pool
	ctx  context.Context
}

var errorMessageTemplate = `
    <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative" role="alert">
        <strong class="font-bold">Oops!</strong>
        <span class="block sm:inline">%s</span>
    </div>`

// Connection pooling
func getDefaultConfig() *PoolConfig {
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

func getConnectionPool(config *PoolConfig) (*pgxpool.Pool, error) {
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

func getUser(c echo.Context) UserAuth {
	return UserAuth{
		c.FormValue("email"),
		c.FormValue("password"),
		c.FormValue("confirm-password"),
		c.FormValue("privacy-consent"),
	}
}

func getUserPwd(user UserAuth, pgContext *PostgresContext) (string, error) {
	var realPwd string
	const query = "SELECT password FROM users WHERE email = $1"
	err := pgContext.pool.QueryRow(pgContext.ctx, query, user.Email).Scan(&realPwd)
	return realPwd, err
}
