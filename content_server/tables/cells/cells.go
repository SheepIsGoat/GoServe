package cells

import (
	"bytes"
	"html/template"
	"log"
)

type TableCell struct {
	Data CellData
}

type CellData interface {
	TemplateName() string
}

func (t *TableCell) RenderCell(tmpl *template.Template) (template.HTML, error) {
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, t.Data.TemplateName(), t.Data)
	if err != nil {
		log.Printf("Error executing cell template: %v\n", err)
		return "", err
	}

	return template.HTML(buf.String()), err
}

type ProfileCell struct {
	Avatar string
	Name   string
	Title  string
}

func (p ProfileCell) TemplateName() string {
	return "tableCell/profile"
}

type StatusCell struct {
	Status string
	Color  string
}

func (p StatusCell) TemplateName() string {
	return "tableCell/status"
}

var StatusColorMap = map[string]string{
	"Approved": "green",
	"Pending":  "orange",
	"Denied":   "red",
	"Expired":  "grey",
}

var ColorCssMap = map[string]string{
	"green":  "text-green-700  bg-green-100  dark:bg-green-700  dark:text-green-100",
	"orange": "text-orange-700 bg-orange-100 dark:bg-orange-600 dark:text-white",
	"red":    "text-red-700    bg-red-100    dark:bg-red-700    dark:text-red-100",
	"grey":   "text-gray-700   bg-gray-100   dark:text-gray-100 dark:bg-gray-700",
	"purple": "text-white transition-colors bg-purple-600 active:bg-purple-600 hover:bg-purple-700 focus:outline-none focus:shadow-outline-purple",
}

type BasicCell struct {
	Val string
}

func (p BasicCell) TemplateName() string {
	return "tableCell/basic"
}
