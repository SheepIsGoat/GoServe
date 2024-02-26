package main

import (
	"encoding/base64"
	"log"
	pg "main/postgres"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserAuth struct {
	Email           string
	Password        string
	ConfirmPassword string
	Consent         string
}

func getUser(c echo.Context) UserAuth {
	return UserAuth{
		c.FormValue("email"),
		c.FormValue("password"),
		c.FormValue("confirm-password"),
		c.FormValue("privacy-consent"),
	}
}

func (hCtx *HandlerContext) validateSignup(user UserAuth) ([]byte, string, bool) {
	var passHash []byte
	notOk := false
	validEmail := govalidator.IsEmail(user.Email)
	if !validEmail {
		return passHash, "Invalid email address", notOk
	}

	if user.Password != user.ConfirmPassword {
		return passHash, "Passwords must match", notOk
	}

	if len(user.Password) < 6 {
		return passHash, "Password must be at least 6 characters", notOk
	}

	if user.Consent != "agree" {
		return passHash, "You must agree to the privacy policy", notOk
	}

	var exists bool
	checkUserSql := `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`
	err := hCtx.PGCtx.Pool.QueryRow(hCtx.PGCtx.Ctx, checkUserSql, user.Email).Scan(&exists)
	if err != nil {
		log.Printf("Existing user check - query failed: %v for email %v. Error: %v", checkUserSql, user.Email, err)
		return passHash, "Internal server error", notOk
	}

	if exists {
		return passHash, "An account with this email already exists", notOk
	}

	passHash, err = bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password for email %v", user.Email)
		return passHash, "Internal server error", notOk
	}
	ok := true
	return passHash, "Success", ok
}

func insertAccount(PGCtx *pg.PostgresContext, user UserAuth, passHash []byte) error {
	encodedPassHash := base64.StdEncoding.EncodeToString(passHash)

	uuid := uuid.New().String()

	sqlStatement := `INSERT INTO users (id, email, password) VALUES ($1, $2, $3)`

	// Execute the SQL statement with the desired values.
	_, err := PGCtx.Pool.Exec(PGCtx.Ctx, sqlStatement, uuid, user.Email, encodedPassHash)
	return err
}

func (hCtx *HandlerContext) authenticateUser() (string, string, int) {
	user := getUser(hCtx.EchoCtx)
	uuid, realPwd, err := getUserPwd(user, hCtx.PGCtx)
	if err != nil {
		log.Printf("QueryRow failed for user %v: %v\n", uuid, err)
		return uuid, "Invalid login credentials", http.StatusUnauthorized
	}
	decodedRealPwd, err := base64.StdEncoding.DecodeString(realPwd)
	if err != nil {
		log.Printf("Failed to b64 decode password for user %v: %v\n", uuid, err)
		return uuid, "Internal server error", http.StatusInternalServerError
	}

	if err := bcrypt.CompareHashAndPassword(decodedRealPwd, []byte(user.Password)); err != nil {
		log.Printf("Incorrect password for user: %v\n", uuid)
		return uuid, "Invalid login credentials", http.StatusUnauthorized
	}
	return uuid, "", http.StatusOK
}

func setCookie(c echo.Context, uuid string) bool {
	expiry := time.Now().Add(time.Hour * 72)
	claims := &jwt.MapClaims{
		"exp":    expiry.Unix(),
		"Issuer": "ResumeSheep",
		"ID":     uuid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("Error signing jwt for user %v\n", uuid)
		return false
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
	return true
}

func getUserPwd(user UserAuth, pgContext *pg.PostgresContext) (string, string, error) {
	var uuid, realPwd string
	const query = "SELECT id, password FROM users WHERE email = $1"
	err := pgContext.Pool.QueryRow(pgContext.Ctx, query, user.Email).Scan(&uuid, &realPwd)
	return uuid, realPwd, err
}
