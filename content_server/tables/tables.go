package tables

import (
	"html/template"
	"log"
	pg "main/postgres"
	page "main/tables/pagination"
	"main/tables/rows"

	"github.com/labstack/echo/v4"
)

type Table[T rows.Row] struct {
	Headers    []string
	TableData  []template.HTML //rows.RowBuilder[T]
	Pagination page.Pagination
}

func (t *Table[T]) CountRows(
	pgContext *pg.PostgresContext,
	rowProcessor rows.RowProcessor[T],
) (int, error) {
	count, err := rowProcessor.Count(pgContext)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (t *Table[T]) RenderTable(
	c echo.Context,
	pgContext *pg.PostgresContext,
	tmpl *template.Template,
	rowProcessor rows.RowProcessor[T],
) error {
	builder := rows.RowBuilder[T]{
		RowProcessor: rowProcessor,
		Tmpl:         tmpl,
	}
	tableData, err := builder.BuildTableData(pgContext, t.Pagination.Config)
	if err != nil {
		log.Printf("Error building table data: %v", err)
		return err
	}
	t.TableData = tableData
	t.Headers = rowProcessor.GetHeaders()
	err = tmpl.ExecuteTemplate(c.Response().Writer, "table", t)
	if err != nil {
		log.Printf("Error executing component template: %v\n", err)
		return err
	}
	return nil
}
