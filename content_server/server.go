package main

import (
	"context"
	"log"

	// "main/tube"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var jwtSecret = []byte("your_jwt_secret")

var StaticPath = "static/public"
var AssetsPath = filepath.Join(StaticPath, "assets")
var TemplatesPath = filepath.Join(StaticPath, "*.html")

func main() {

	// videoID := "kHuX4huT1sc"

	// tube.Download(videoID)

	e := echo.New()

	e.HTTPErrorHandler = customHTTPErrorHandler
	e.Pre(middleware.AddTrailingSlash())
	e.Use(middleware.Logger())

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob(TemplatesPath)),
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

		c.Response().Header().Set("HX-Redirect", "/app/")
		return c.NoContent(http.StatusOK)
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
	app.Use(JWTFromCookie())
	app.Use(jwtClaimsMiddleware(strClaimsValidation, f64ClaimsValidation))
	app.Use(PgxPoolMiddleware(pool))

	app.GET("/", func(c echo.Context) error {
		// pgContext := PostgresContext{pool, context.Background()}
		data := map[string]interface{}{}
		return RenderTemplate(c, http.StatusOK, "app.html", "dashboard", data)
	})

	serveFile := func(c echo.Context, filename string) error {
		err := c.File(filepath.Join(StaticPath, filename))
		if err != nil {
			log.Printf("Error serving file %s: %v", filename, err)
			return err
		}
		return nil
	}

	app.GET("/dashboard/", func(c echo.Context) error {
		return serveFile(c, "dashboard.html")
	}).Name = "index"

	app.GET("/forms/", func(c echo.Context) error {
		return serveFile(c, "forms.html")
	}).Name = "index"

	app.GET("/charts/", func(c echo.Context) error {
		return serveFile(c, "charts.html")
	}).Name = "index"

	app.GET("/cards/", func(c echo.Context) error {
		return serveFile(c, "cards.html")
	}).Name = "index"

	app.GET("/buttons/", func(c echo.Context) error {
		return serveFile(c, "buttons.html")
	}).Name = "index"

	app.GET("/modals/", func(c echo.Context) error {
		return serveFile(c, "modals.html")
	}).Name = "index"

	app.GET("/tables/", func(c echo.Context) error {
		return serveFile(c, "tables.html")
	}).Name = "index"

	e.Static("/assets", AssetsPath)
	app.Static("/assets", AssetsPath)

	e.Logger.Fatal(e.Start(":8080"))
}
