package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

// TemplateRenderer is a custom html/template renderer for Echo framework.
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document.
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func appTemplate(html string) string {
	return fmt.Sprintf(`{{ define "content" }}{{ template "%s.html" . }}{{ end }}`, html)
}

func RenderTemplate(
	c echo.Context,
	status int,
	outerHtml string,
	innerHtml string,
	data map[string]interface{},
) error {
	// Simple template rendering with an outer and inner template
	tmpl, err := template.ParseGlob(TemplatesPath)
	if err != nil {
		log.Printf("Failed to parse inner template glob for html %v. err: %v\n", innerHtml, err)
		return err
	}

	tmpl, err = tmpl.Parse(appTemplate(innerHtml))
	if err != nil {
		log.Printf("Failed to render template for html %v. err: %v\n", innerHtml, err)
		return err
	}

	return tmpl.ExecuteTemplate(c.Response().Writer, outerHtml+".html", data)
}

func serveFile(c echo.Context, filename string) error {
	err := c.File(filepath.Join(StaticPath, filename))
	if err != nil {
		log.Printf("Error serving file %s: %v", filename, err)
		return err
	}
	log.Printf("I'm serving the file %s", filename)
	return nil
}
