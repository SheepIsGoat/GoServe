package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
)

type PoolConfig struct {
	ConnectionString string
	MaxConns         int32
	MinConns         int32
	ConnectTimeout   time.Duration
}

func GetDefaultConfig() *PoolConfig {
	return &PoolConfig{
		ConnectionString: "postgresql://postgres:mysecretpassword@localhost:5432/postgres",
		MaxConns:         10,
		MinConns:         5,
		ConnectTimeout:   time.Second * 5,
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

func loginEndpoint(c echo.Context, pool *pgxpool.Pool) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	var realPwd string
	const query = "SELECT password FROM users WHERE username = $1"
	err := pool.QueryRow(context.Background(), query, username).Scan(&realPwd)
	if err != nil {
		log.Printf("QueryRow failed for user %v: %v", username, err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User does not exist"})
	}
	decodedRealPwd, err := base64.StdEncoding.DecodeString(realPwd)
	if err != nil {
		log.Printf("Failed to decode pw from database for user: %v", username)
		return echo.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword(decodedRealPwd, []byte(password)); err != nil {
		log.Printf("Incorrect password for user: %v", username)
		return echo.ErrUnauthorized
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": t,
	})
}

func accessible(c echo.Context) error {
	return c.String(http.StatusOK, "Accessible")
}

func restricted(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return c.String(http.StatusOK, "Welcome "+name+"!")
}

func main() {
	e := echo.New()

	config := GetDefaultConfig()
	pool, err := GetConnectionPool(config)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	e.POST("/login", func(c echo.Context) error {
		return loginEndpoint(c, pool)
	})

	e.GET("/", accessible)

	r := e.Group("/restricted")
	r.Use(middleware.JWT([]byte("secret")))
	r.GET("", restricted)

	e.Start(":8080")
}
