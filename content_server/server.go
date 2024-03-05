package main

import (
	"context"
	"log"

	"html/template"
	pg "main/postgres"
	tp "main/templating"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var jwtSecret = []byte("your_jwt_secret")

var StaticPath = "static/public"
var AssetsPath = filepath.Join(StaticPath, "assets")
var TemplatesPath = StaticPath + "/*/*.html"

func main() {

	// videoID := "kHuX4huT1sc"
	// tube.Download(videoID)

	initFilesystem()

	// setup
	e := echo.New()

	tmpl, err := tp.GetTmpl(StaticPath)
	e.HTTPErrorHandler = customHTTPErrorHandler(tmpl)
	e.Pre(middleware.AddTrailingSlash())
	e.Use(middleware.Logger())

	renderer := &tp.TemplateRenderer{
		Templates: template.Must(tmpl, err),
	}

	e.Renderer = renderer

	config := pg.GetDefaultConfig()
	pool, err := pg.GetConnectionPool(config)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	// public endpoints
	e.GET("/login/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login", map[string]interface{}{})
	}).Name = "login"

	e.POST("/login/", func(c echo.Context) error {
		hCtx := HandlerContext{c, &pg.PostgresContext{pool, context.Background()}}
		err := hCtx.loginEndpoint()
		if err != nil {
			return err
		}

		c.Response().Header().Set("HX-Redirect", "/app/")
		return c.NoContent(http.StatusOK)
	})

	e.GET("/create-account/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "create-account", map[string]interface{}{})
	}).Name = "create-account"

	e.POST("/create-account/", func(c echo.Context) error {
		hCtx := HandlerContext{c, &pg.PostgresContext{pool, context.Background()}}
		return hCtx.createAccount()
	})

	// private app group
	app := e.Group("/app")
	app.Use(JWTFromCookie())
	app.Use(jwtClaimsMiddleware(strClaimsValidation, f64ClaimsValidation))
	app.Use(PgxPoolMiddleware(pool))

	app.GET("/", func(c echo.Context) error {
		data := map[string]interface{}{}
		return tp.RenderTemplate(c, tmpl, "app", "dashboard", data)
	})

	// page components
	app.GET("/dashboard/", func(c echo.Context) error {
		return tp.ServeFile(c, tmpl, "dashboard")
	}).Name = "index"

	app.GET("/files/", func(c echo.Context) error {
		return tp.ServeFile(c, tmpl, "files")
	}).Name = "index"

	app.GET("/forms/", func(c echo.Context) error {
		return tp.ServeFile(c, tmpl, "forms")
	}).Name = "index"

	app.GET("/charts/", func(c echo.Context) error {
		return tp.ServeFile(c, tmpl, "charts")
	}).Name = "index"

	app.GET("/cards/", func(c echo.Context) error {
		return tp.ServeFile(c, tmpl, "cards")
	}).Name = "index"

	app.GET("/buttons/", func(c echo.Context) error {
		return tp.ServeFile(c, tmpl, "buttons")
	}).Name = "index"

	app.GET("/modals/", func(c echo.Context) error {
		return tp.ServeFile(c, tmpl, "modals")
	}).Name = "index"

	app.GET("/tables/", func(c echo.Context) error {
		return tp.ServeFile(c, tmpl, "tables")
	}).Name = "index"

	// endpoints
	app.POST("/files/upload/", func(c echo.Context) error {
		hCtx := HandlerContext{c, &pg.PostgresContext{pool, context.Background()}}
		return FileUpload(hCtx)
	}).Name = "index"

	app.POST("/files/delete/", func(c echo.Context) error {
		hCtx := HandlerContext{c, &pg.PostgresContext{pool, context.Background()}}
		return FileDelete(hCtx)
	}).Name = "index"

	app.GET("/table/", func(c echo.Context) error {
		hCtx := HandlerContext{c, &pg.PostgresContext{pool, context.Background()}}
		return Table(&hCtx, tmpl)
	}).Name = "index"

	app.GET("/charts/pie/", func(c echo.Context) error {
		hCtx := HandlerContext{c, &pg.PostgresContext{pool, context.Background()}}
		log.Println("Hitting table endpoint")
		// return tables.RenderTable(c, tmpl)
		return hCtx.PieChart(tmpl)
	}).Name = "index"

	// static assets
	e.Static("/assets", AssetsPath)
	app.Static("/assets", AssetsPath)

	// start server
	e.Logger.Fatal(e.Start(":8080"))
}
