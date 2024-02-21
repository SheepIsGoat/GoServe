package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	// echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/asaskevich/govalidator"
)

// Output is a generic struct for handling results and errors.
type Output[T any] struct {
	Result T
	Err    error
}

// ValidatorFunc is a type for validation functions.
type ValidatorFunc[T any] func(T) (error)

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
                if err == http.ErrNoCookie {
                    return c.String(http.StatusUnauthorized, "No token in cookie")
                }
                return err
            }

            tokenString := cookie.Value
            token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
                // Don't forget to validate the alg is what you expect:
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
                }
                return jwtSecret, nil
            })

            if err != nil {
                return c.String(http.StatusUnauthorized, "Invalid token")
            }

            if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
                c.Set("user", token)
                return next(c)
            } else {
                return c.String(http.StatusUnauthorized, "Invalid token")
            }
        }
    }
}


// jwtClaimsMiddleware handles the JWT claims validation using a ValidationMap.
func jwtClaimsMiddleware(sMap ValidationMap[string], fMap ValidationMap[float64]) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			fmt.Printf("c.Get(\"user\"): %v\n", c.Get("user"))




			unJwt, ok := c.Get("user").(*jwt.Token)
			// unJwt, ok := c.Cookie("auth_token")
			if !ok {
				return c.String(http.StatusUnauthorized, "Invalid token")
			}

			claims, ok := unJwt.Claims.(jwt.MapClaims)
			if !ok {
				return c.String(http.StatusUnauthorized, "Invalid claims format")
			}

			for claim, entry := range sMap {
				claimValue, ok := claims[claim].(string)

				if !ok {
					if entry.Required {
						return c.String(http.StatusUnauthorized, fmt.Sprintf("%s not found in token", claim))
					}
					continue // Skip validation for optional, missing claims
				}

				err := entry.Func(claimValue)
				if err != nil {
					return c.String(http.StatusUnauthorized, err.Error())
				}
				c.Set(claim, claimValue)
			}

			for claim, entry := range fMap {
				claimValue, ok := claims[claim].(float64)

				if !ok {
					if entry.Required {
						return c.String(http.StatusUnauthorized, fmt.Sprintf("%s not found in token", claim))
					}
					continue // Skip validation for optional, missing claims
				}

				err := entry.Func(claimValue)
				if err != nil {
					return c.String(http.StatusUnauthorized, err.Error())
				}
				c.Set(claim, claimValue)
			}

			return next(c)
		}
	}
}


var strClaimsValidation = ValidationMap[string]{
	"Email": {Func: validateEmail, Required: true},
	"Issuer":     {Func: validateIssuer, Required: true},
}

var f64ClaimsValidation = ValidationMap[float64]{
	"exp":     {Func: validateExpiry, Required: true},
}



func validateEmail(email string) (error) {
	limit := 32
	if len(email) > limit {
		message := fmt.Sprintf("Email too long. %v > %v", len(email), limit)
		return NewClaimsError("email", message)
	}
	if !govalidator.IsEmail(email) {
		message := fmt.Sprintf("Not a valid email address: %v", email)
		return NewClaimsError("email", message)
	}
	return nil
}

func validateExpiry(exp float64) (error) {
	expTime := time.Unix(int64(exp), 0)
	if time.Now().After(expTime) {
		message := fmt.Sprintf("JWT expired: %v < %v", expTime, time.Now())
		return NewClaimsError("exp", message)
	}
	return nil
}

func validateIssuer(issuer string) (error) {
	if issuer != "ResumeSheep" {
		message := fmt.Sprintf("Invalid JWT issuer: %v != %v", issuer, "ResumeSheep")
		return NewClaimsError("Issuer", message)
	}
	return nil
}


// // Time format to use for parsing
var timeFormat = "2006-01-02T15:04:05.999999999-07:00"
//
// // validateTime validates if the time is within a 24-hour range.
// func validateTime(timeStr string) (string, error) {
// 	parsedTime, err := time.Parse(timeFormat, timeStr)
// 	if err != nil {
// 		message := "Could not parse time"
// 		return timeStr, NewClaimsError(timeStr, message)
// 	}
// 	if !isTimeWithin24Hours(parsedTime) {
// 		message := "Time out of bounds"
// 		return timeStr, NewClaimsError(timeStr, message)
// 	}
// 	return timeStr, nil
// }

// // isTimeWithin24Hours checks if the time is within the past 24 hours.
// func isTimeWithin24Hours(parsedTime time.Time) bool {
// 	currentTime := time.Now()
// 	twentyFourHoursAgo := currentTime.Add(-24 * time.Hour)
// 	twentyFourHoursFromNow := currentTime.Add(24 * time.Hour)

// 	return parsedTime.After(twentyFourHoursAgo) && parsedTime.Before(twentyFourHoursFromNow)
// }

// PgxPoolMiddleware injects a Postgres connection pool into the Echo context.
func PgxPoolMiddleware(pool *pgxpool.Pool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := context.Background()
			pgContext := &PostgresContext{
				pool: pool,
				ctx:  ctx,
			}

			c.Set("pgContext", pgContext)

			return next(c)
		}
	}
}
