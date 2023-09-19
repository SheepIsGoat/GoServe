package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"time"
	"errors"

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

func getDefaultConfig() *PoolConfig {
	return &PoolConfig{
		ConnectionString: "postgresql://postgres:mysecretpassword@localhost:5432/server_db",
		MaxConns:         10,
		MinConns:         5,
		ConnectTimeout:   time.Second * 5,
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

func createUser(c echo.Context, pool *pgxpool.Pool) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password for username %v", username)
		return err
	}

	encodedPassHash := base64.StdEncoding.EncodeToString(passHash)

	sqlStatement := `INSERT INTO users (username, password) VALUES ($1, $2)`

    // Execute the SQL statement with the desired values.
    _, err = pool.Exec(context.Background(), sqlStatement, username, encodedPassHash)
	if err != nil {
		log.Printf("Failed to create user with username %v. Error: %v", username, err)
	}
	return err
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

type ParentContentStruct struct {
	Username string
	NewContentStruct
}

func content(c echo.Context, pool *pgxpool.Pool, contentINF ParentContentStruct, contentFunc func (*pgxpool.Pool, string, ParentContentStruct) ([]map[string]string, error)) (error) {
    unJwt, ok := c.Get("user").(*jwt.Token)
    if !ok {
        return c.String(http.StatusUnauthorized, "Invalid token")
    }

    claims, ok := unJwt.Claims.(jwt.MapClaims)
    if !ok {
        return c.String(http.StatusUnauthorized, "Invalid claims format")
    }

    username, ok := claims["username"].(string)
    if !ok {
        return c.String(http.StatusUnauthorized, "Name not found in token")
    }

	// type jsonData struct {
	// 	time string
	// }
	// var content jsonData
	if err := c.Bind(&contentINF); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	// timeStr := content.time

	// input := map[string]string{"username": username, "time": timeStr}
	content, err := contentFunc(pool, username, contentINF)
	if err != nil {
		return c.String(500, err.Error())
	}
    return c.JSON(http.StatusOK, content)
}

type NewContentStruct struct {
	Time string
}

func newContent(pool *pgxpool.Pool, username string, timeStruct ParentContentStruct) ([]map[string]string, error) {
	
	
	timeStr := timeStruct.NewContentStruct.Time
	if timeStr == "" {
		return nil, errors.New("Time string is empty")
	}
	
	timeFormat := "2006-01-02T15:04:05"
	timeObj, err := time.Parse(timeFormat, timeStr)
	if err != nil {
		return nil, errors.New("Could not parse time " + timeStr)
	}

	query := "SELECT title, summary, background_url FROM content WHERE datetime>$1"
	rows, err := pool.Query(context.Background(), query, timeObj)
	if err != nil {
		return nil, errors.New("Query failed: " + err.Error())
	}
	defer rows.Close()


	var result []map[string]string
	for rows.Next() {
		var title, summary, background string
		err := rows.Scan(&title, &summary, &background)
		if err != nil {
			log.Fatalf("Row scan failed: %v\n", err)
		}
		row := map[string]string {
			"title": title,
			"summary": summary,
			"background": background,
		}
		
		result = append(result, row)
	}
	
	return result, nil
}


func accessible(c echo.Context) error {
	return c.String(http.StatusOK, "Accessible")
}

func restricted(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	return c.String(http.StatusOK, "Welcome "+username+"!")
}

func main() {
	e := echo.New()

	config := getDefaultConfig()
	pool, err := getConnectionPool(config)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	e.POST("/login", func(c echo.Context) error {
		return loginEndpoint(c, pool)
	})

	e.POST("/create_user", func(c echo.Context) error {
		return createUser(c, pool)
	})

	r := e.Group("/new_content")
	r.Use(middleware.JWT([]byte("secret")))
	r.GET("", func(c echo.Context) error{
		return content(c, pool, ParentContentStruct{}, newContent)
	})

	ra := e.Group("/new_content_p")
	ra.Use(middleware.JWT([]byte("secret")))
	ra.POST("", func(c echo.Context) error{
		return content(c, pool, ParentContentStruct{}, newContent)
	})


	e.GET("/", accessible)

	// r := e.Group("/restricted")
	// r.Use(middleware.JWT([]byte("secret")))
	// r.GET("", restricted)

	e.Start(":8080")
}
