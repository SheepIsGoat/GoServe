package main

import (
	"context"
	"log"
	// "main/tube"
	"html/template"
	"net/http"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	// "github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your_jwt_secret")

func main() {

	// videoID := "kHuX4huT1sc"

	// tube.Download(videoID)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Pre(middleware.AddTrailingSlash())

    renderer := &TemplateRenderer{
        templates: template.Must(template.ParseGlob("static/public/*.html")),
    }

    e.Renderer = renderer

	config := getDefaultConfig()
	pool, err := getConnectionPool(config)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()


	e.GET("/login/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", map[string]interface{}{})
	}).Name = "login"

	e.POST("/login/", func(c echo.Context) error {
		user := getUser(c)
		pgContext := PostgresContext{pool, context.Background()}
		err := loginEndpoint(c, user, &pgContext)
		if err != nil {
			return err
		}

    	return c.Redirect(http.StatusFound, "/app/")
	})


	e.GET("/create-account/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "create-account.html", map[string]interface{}{})
	}).Name = "create-account"

	e.POST("/create-account/", func(c echo.Context) error {
		user := getUser(c)
		pgContext := PostgresContext{pool, context.Background()}
		return createAccount(c, user, &pgContext)
	})

	app := e.Group("/app")
	app.Use(echojwt.JWT(jwtSecret))
	app.Use(JWTFromCookie())
	app.Use(jwtClaimsMiddleware(strClaimsValidation, f64ClaimsValidation))
	app.Use(PgxPoolMiddleware(pool))
	app.GET("/", func(c echo.Context) error {
		// pgContext := PostgresContext{pool, context.Background()}
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{})
	})


	r := e.Group("/content")
	r.Use(echojwt.JWT(jwtSecret))
	r.Use(jwtClaimsMiddleware(strClaimsValidation, f64ClaimsValidation))
	r.Use(PgxPoolMiddleware(pool))
	r.GET("/new", func(c echo.Context) error {
		pgContext := PostgresContext{pool, context.Background()}
		return newContentHandler(c, pgContext)
	})

	// e.Static("/app", "static/public")
	e.Static("/app/assets", "static/public/assets")


	e.Logger.Fatal(e.Start(":8080"))
}
