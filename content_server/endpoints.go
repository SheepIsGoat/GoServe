package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func createAccount(c echo.Context, pgContext *PostgresContext) error {
	user := getUser(c)
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

func loginEndpoint(c echo.Context, pgContext *PostgresContext) error {
	user := getUser(c)
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
		"exp":    expiry.Unix(),
		"Issuer": "ResumeSheep",
		"Email":  user.Email,
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

func upload(c echo.Context, pgContext *PostgresContext) {
	user := getUser(c)

	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
}
