package main

import (
	"context"
	"fmt"
	"html/template"
	pg "main/postgres"
	"main/templating"
	"net/http"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
)

// Output is a generic struct for handling results and errors.
type Output[T any] struct {
	Result T
	Err    error
}

// ValidatorFunc is a type for validation functions.
type ValidatorFunc[T any] func(T) error

// ValidatorEntry is a struct to handle validation function and its required status.
type ValidatorEntry[T any] struct {
	Func     ValidatorFunc[T]
	Required bool
}

// ValidationMap is a map that ties claim names to their respective ValidatorEntry.
type ValidationMap[T any] map[string]ValidatorEntry[T]

// NewClaimsError constructs a new error message for JWT claims.
func NewClaimsError(key string, message string) error {
	return fmt.Errorf("JWT claims error. Key: %v. Message: %v", key, message)
}

func JWTFromCookie() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("auth_token")
			if err != nil {
				return c.Redirect(http.StatusFound, "/login")
			}

			tokenString := cookie.Value
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return jwtSecret, nil
			})

			if err != nil {
				return c.Redirect(http.StatusFound, "/login")
			}

			if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				c.Set("user", token)
				return next(c)
			} else {
				return c.Redirect(http.StatusFound, "/login")
			}
		}
	}
}

// jwtClaimsMiddleware handles the JWT claims validation using a ValidationMap.
func jwtClaimsMiddleware(sMap ValidationMap[string], fMap ValidationMap[float64]) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			unJwt, ok := c.Get("user").(*jwt.Token)
			// unJwt, ok := c.Cookie("auth_token")
			if !ok {
				return c.Redirect(http.StatusFound, "/login")
			}

			claims, ok := unJwt.Claims.(jwt.MapClaims)
			if !ok {
				return c.Redirect(http.StatusFound, "/login")
			}

			for claim, entry := range sMap {
				claimValue, ok := claims[claim].(string)

				if !ok {
					if entry.Required {
						return c.Redirect(http.StatusFound, "/login")
					}
					continue // Skip validation for optional, missing claims
				}

				err := entry.Func(claimValue)
				if err != nil {
					return c.Redirect(http.StatusFound, "/login")
				}
				c.Set(claim, claimValue)
			}

			for claim, entry := range fMap {
				claimValue, ok := claims[claim].(float64)

				if !ok {
					if entry.Required {
						return c.Redirect(http.StatusFound, "/login")
					}
					continue // Skip validation for optional, missing claims
				}

				err := entry.Func(claimValue)
				if err != nil {
					return c.Redirect(http.StatusFound, "/login")
				}
				c.Set(claim, claimValue)
			}

			return next(c)
		}
	}
}

var strClaimsValidation = ValidationMap[string]{
	"ID":     {Func: validateUUID, Required: true},
	"Issuer": {Func: validateIssuer, Required: true},
}

var f64ClaimsValidation = ValidationMap[float64]{
	"exp": {Func: validateExpiry, Required: true},
}

func validateUUID(uuid string) error {
	if len(uuid) != 36 {
		message := fmt.Sprintf("Invalid ID length: %v. Expected length is 36.", len(uuid))
		return NewClaimsError("ID", message)
	}

	uidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uidRegex.MatchString(uuid) {
		message := fmt.Sprintf("Invalid UUID format: %v", uuid)
		return NewClaimsError("ID", message)
	}

	return nil
}

// func validateEmail(email string) error {
// 	limit := 32
// 	if len(email) > limit {
// 		message := fmt.Sprintf("Email too long. %v > %v", len(email), limit)
// 		return NewClaimsError("email", message)
// 	}
// 	if !govalidator.IsEmail(email) {
// 		message := fmt.Sprintf("Not a valid email address: %v", email)
// 		return NewClaimsError("email", message)
// 	}
// 	return nil
// }

func validateExpiry(exp float64) error {
	expTime := time.Unix(int64(exp), 0)
	if time.Now().After(expTime) {
		message := fmt.Sprintf("JWT expired: %v < %v", expTime, time.Now())
		return NewClaimsError("exp", message)
	}
	return nil
}

func validateIssuer(issuer string) error {
	if issuer != "ResumeSheep" {
		message := fmt.Sprintf("Invalid JWT issuer: %v != %v", issuer, "ResumeSheep")
		return NewClaimsError("Issuer", message)
	}
	return nil
}

// PgxPoolMiddleware injects a Postgres connection pool into the Echo context.
func PgxPoolMiddleware(pool *pgxpool.Pool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := context.Background()
			pgContext := &pg.PostgresContext{
				Pool: pool,
				Ctx:  ctx,
			}

			c.Set("pgContext", pgContext)

			return next(c)
		}
	}
}

func customHTTPErrorHandler(tmpl *template.Template) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		c.Logger().Error(err)

		_, cookieErr := c.Cookie("auth_token")
		if cookieErr != nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		messages := map[int]string{
			404: "Page not found",
			400: "Bad request",
			401: "Unauthorized",
			403: "Forbidden",
			500: "Internal server error",
			502: "Bad gateway",
			503: "Service unavailable",
		}

		data := map[string]interface{}{
			"ErrorCode":    code,
			"ErrorMessage": messages[code],
		}
		if err := templating.RenderTemplate(c, tmpl, "app", "error", data); err != nil {
			c.Logger().Error(err)
		}
	}
}
