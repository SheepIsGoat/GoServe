package rows

import (
	"html/template"
	"log"
	pg "main/postgres"
	"main/tables/pagination"
	"main/templating/components"
	"strings"
)

type RowBuilder[T Row] struct {
	RowProcessor RowProcessor[T]
	Tmpl         *template.Template
}

type Row interface {
}

type RowProcessor[T Row] interface {
	Count(*pg.PostgresContext) (int, error)
	QuerySQLToStructArray(*pg.PostgresContext, pagination.PaginConfig) ([]T, error)
	BuildRowCells(T) []components.DivComponent
	GetHeaders() []string
}

func (rb *RowBuilder[T]) RenderRow(row T) (template.HTML, error) {
	// requires tr.RowData to be populated
	// returns a single row rendered as html
	var renderedCells strings.Builder

	for _, cell := range rb.RowProcessor.BuildRowCells(row) {
		renderedHTML, err := cell.RenderComponent(rb.Tmpl)
		if err != nil {
			log.Printf("Could not render table cell: %v", err)
			return "", err
		}
		renderedCells.WriteString(string(renderedHTML))
	}

	return template.HTML(renderedCells.String()), nil
}

func (rb *RowBuilder[T]) BuildTableData(pgContext *pg.PostgresContext, pagination pagination.PaginConfig) ([]template.HTML, error) {
	rows := []template.HTML{}
	rawData, err := rb.RowProcessor.QuerySQLToStructArray(pgContext, pagination)
	if err != nil {
		return rows, err
	}
	for _, rawRow := range rawData {
		htmlRow, err := rb.RenderRow(rawRow)
		if err != nil {
			return rows, err
		}
		rows = append(rows, htmlRow)
	}
	return rows, nil
}
