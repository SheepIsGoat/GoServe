package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	
	"github.com/asaskevich/govalidator"
)

// Structs
type PoolConfig struct {
	ConnectionString string
	MaxConns         int32
	MinConns         int32
	ConnectTimeout   time.Duration
}

type UserAuth struct {
	Email string
	Password string
	ConfirmPassword string
	Consent string
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


// Authentication
func createAccount(c echo.Context, user UserAuth, pgContext *PostgresContext) error {

	validEmail := govalidator.IsEmail(user.Email)
	if !validEmail {
		fmt.Printf("User email is not valid: %v\n", user.Email)
		errorMessage := fmt.Sprintf(errorMessageTemplate, "Invalid email address")
		return c.HTML(http.StatusOK, errorMessage)
	}

	if user.Password != user.ConfirmPassword {
		errorMessage := fmt.Sprintf(errorMessageTemplate, "Passwords must match")
		return c.HTML(http.StatusOK, errorMessage)
	}

	if len(user.Password) < 6 {
		errorMessage := fmt.Sprintf(errorMessageTemplate, "Password must be at least 6 characters")
		return c.HTML(http.StatusOK, errorMessage)
	}

	if user.Consent != "agree" {
		errorMessage := fmt.Sprintf(errorMessageTemplate, "You must agree to the privacy policy")
		return c.HTML(http.StatusOK, errorMessage)
	}

	var exists bool
	checkUserSql := `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`
	err := pgContext.pool.QueryRow(pgContext.ctx, checkUserSql, user.Email).Scan(&exists)
	if err != nil {
		log.Printf("Failed to check if user exists with email %v. Error: %v", user.Email, err)
		return err
	}

	if exists {
		errorMessage := fmt.Sprintf(errorMessageTemplate, "An account with this email already exists")
		return c.HTML(http.StatusOK, errorMessage)
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password for email %v", user.Email)
		return err
	}

	encodedPassHash := base64.StdEncoding.EncodeToString(passHash)

	sqlStatement := `INSERT INTO users (email, password) VALUES ($1, $2)`

	// Execute the SQL statement with the desired values.
	_, err = pgContext.pool.Exec(pgContext.ctx, sqlStatement, user.Email, encodedPassHash)
	if err != nil {
		log.Printf("Failed to create user with email %v. Error: %v", user.Email, err)
	}

	log.Printf("Succesfully created user with email %v", user.Email)
	return err
}

func loginEndpoint(c echo.Context, user UserAuth, pgContext *PostgresContext) error {
	realPwd, err := getUserPwd(user, pgContext)
	if err != nil {
		log.Printf("QueryRow failed for user %v: %v\n", user.Email, err)
		errorMessage := fmt.Sprintf(errorMessageTemplate, "Invalid login credentials")
		return c.HTML(http.StatusOK, errorMessage)
		// return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User does not exist"})
	}
	decodedRealPwd, err := base64.StdEncoding.DecodeString(realPwd)
	if err != nil {
		log.Printf("Failed to decode pw from database for user: %v\n", user.Email)
		return echo.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword(decodedRealPwd, []byte(user.Password)); err != nil {
		log.Printf("Incorrect password for user: %v\n", user.Email)
		return echo.ErrUnauthorized
	}

	
	expiry := time.Now().Add(time.Hour * 72)
	claims := &jwt.MapClaims{
		"exp": expiry.Unix(),
		"Issuer": "ResumeSheep",
		"Email": user.Email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString(jwtSecret)
	if err != nil {
		fmt.Printf("Error signing jwt for user %v\n", user.Email)
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:     "auth_token",
		Value:    t,
		Expires:  expiry,
		HttpOnly: true,
		Path:     "/",
		Secure:   false, // Set to true if using HTTPS, recommended for security
		SameSite: http.SameSiteStrictMode,
	})

	fmt.Printf("Successful login for user %v\n", user.Email)
	return nil //c.Redirect(http.StatusFound, "/app")
}

func getUserPwd(user UserAuth, pgContext *PostgresContext) (string, error) {
	var realPwd string
	const query = "SELECT password FROM users WHERE email = $1"
	err := pgContext.pool.QueryRow(pgContext.ctx, query, user.Email).Scan(&realPwd)
	return realPwd, err
}
