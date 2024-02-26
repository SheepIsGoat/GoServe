package templating

import (
	"bytes"
	"html/template"
	"main/arithmetic"
	"strings"
	"unicode"

	"io"
	"log"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

// TemplateRenderer is a custom html/template renderer for Echo framework.
type TemplateRenderer struct {
	Templates *template.Template
}

// Render renders a template document.
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.Templates.ExecuteTemplate(w, name, data)
}

func GetTmpl(staticPath string) (*template.Template, error) {
	tmpl := template.New("base").Funcs(funcMap)

	// Define patterns for both root and 'pages' subdirectory
	patterns := []string{
		staticPath + "/*.html",
		staticPath + "/pages/*.html",
		staticPath + "/components/*/*.html",
	}

	// Range over the patterns and parse each
	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			log.Printf("Failed to glob for pattern %s: %v\n", pattern, err)
			return tmpl, err
		}

		if len(files) > 0 {
			tmpl, err = tmpl.ParseFiles(files...)
			if err != nil {
				log.Printf("Failed to parse templates for pattern %s: %v\n", pattern, err)
				return tmpl, err
			}
		}
	}
	return tmpl, nil
}

func RenderTemplate(
	c echo.Context,
	tmpl *template.Template,
	outer string,
	component string,
	data map[string]interface{},
) error {
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, component, data)
	if err != nil {
		log.Printf("Error executing component template: %v\n", err)
		return err
	}

	data["content"] = template.HTML(buf.String())

	err = tmpl.ExecuteTemplate(c.Response().Writer, outer, data)
	if err != nil {
		log.Printf("Error executing main template: %v\n", err)
		return err
	}

	return nil
}

func ServeFile(c echo.Context, tmpl *template.Template, filename string) error {
	err := tmpl.ExecuteTemplate(c.Response().Writer, filename, nil)
	if err != nil {
		log.Printf("Error serving file %s: %v", filename, err)
		return err
	}
	return nil
}

func toHTMLID(s string) string {
	// Replace spaces with hyphens
	id := strings.ReplaceAll(s, " ", "-")

	// Remove any non-alphanumeric, non-hyphen characters
	id = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			return r
		}
		return -1
	}, id)

	// Ensure the id starts with a letter (prepend 'id-' if not)
	if len(id) > 0 && !unicode.IsLetter(rune(id[0])) {
		id = "id-" + id
	}

	return id
}

var funcMap = template.FuncMap{
	"adduint32": arithmetic.Add[uint32],
	"subuint32": arithmetic.Sub[uint32],
	"toHTMLID":  toHTMLID,
}
