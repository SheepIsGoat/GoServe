package charts

import (
	"encoding/json"
	"html/template"
	"log"
	pg "main/postgres"
	"main/templating/components"

	"github.com/labstack/echo/v4"
)

type (
	ChartBuilder[query any, raw any, display any] struct {
		// Used to create a ChartDisplay object of the specified type, for rendering charts.js
		ChartProcessor ChartProcessor[query, raw, display]
		Tmpl           *template.Template
	}

	ChartProcessor[query any, raw any, display any] interface {
		FetchData(*pg.PostgresContext, query) ([]raw, error)
		PopulateDisplay([]raw) (*display, error)
		_validateQueryCast(query) ChartQueryParams
		_validateRawCast(raw) ChartRawData
		_validateDisplayCast(*display) ChartDisplay
	}

	ChartQueryParams interface {
		_isQueryParams() bool
	}

	ChartRawData interface {
		_isChart() bool
	}

	ChartDisplay interface {
		components.DivData //TemplateName() string
		Init()
		_isDisplay() bool
	}
)

func (cb ChartBuilder[query, raw, display]) RenderChart(
	c echo.Context,
	pgContext *pg.PostgresContext,
	tmpl *template.Template,
	chartQuery query,
) error {
	data, err := cb.ChartProcessor.FetchData(pgContext, chartQuery)

	if err != nil {
		log.Printf("Failed to fetch chart data for %T: %v", chartQuery, err)
		return err
	}

	_charDisp, err := cb.ChartProcessor.PopulateDisplay(data)
	if err != nil {
		log.Printf("Failed to populate display for %T: %v", chartQuery, err)
		return err
	}
	chartDisplay := cb.ChartProcessor._validateDisplayCast(_charDisp)

	jsonData, err := json.Marshal(chartDisplay)
	if err != nil {
		log.Printf("Failed to cast chart input as json for %T: %v", chartQuery, err)
		return err
	}

	templateName := chartDisplay.TemplateName()

	// var buf bytes.Buffer
	// err = tmpl.ExecuteTemplate(&buf, templateName, jsonData)
	// if err != nil {
	// 	log.Printf("Error executing chart template for %s: %v\n", templateName, err)
	// 	return output, err
	// }

	// return template.HTML(buf.String()), err

	err = tmpl.ExecuteTemplate(c.Response().Writer, templateName, template.JS(jsonData))
	if err != nil {
		log.Printf("Error executing chart template: %v\n", err)
		return err
	}
	return nil
}
