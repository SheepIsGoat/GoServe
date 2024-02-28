package components

import (
	"bytes"
	"html/template"
	"log"
)

type DivComponent struct {
	Data DivData
}

type DivData interface {
	TemplateName() string
}

func (t *DivComponent) RenderComponent(tmpl *template.Template) (template.HTML, error) {
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, t.Data.TemplateName(), t.Data)
	if err != nil {
		log.Printf("Error executing cell template: %v\n", err)
		return "", err
	}

	return template.HTML(buf.String()), err
}
