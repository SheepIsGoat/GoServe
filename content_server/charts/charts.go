package charts

import (
	"encoding/json"
	"fmt"
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
		GetLegend() (htmlLegend, error)
		_isDisplay() bool
	}

	htmlLegend struct {
		Title  string
		Labels []coloredItem
	}

	coloredItem struct {
		Label    string
		HexColor string
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
	fmt.Printf("Got chart return data: %+v\n", data)

	_charDisp, err := cb.ChartProcessor.PopulateDisplay(data)
	if err != nil {
		log.Printf("Failed to populate display for %T: %v", chartQuery, err)
		return err
	}
	chartDisplay := cb.ChartProcessor._validateDisplayCast(_charDisp)
	fmt.Printf("Populated chart display: %+v\n", chartDisplay)

	jsonData, err := json.Marshal(chartDisplay)
	if err != nil {
		log.Printf("Failed to cast chart input as json for %T: %v", chartQuery, err)
		return err
	}
	fmt.Printf("Jsonified chart display: %v\n", string(jsonData))

	legend, err := chartDisplay.GetLegend()
	if err != nil {
		return err
	}

	templateInput := make(map[string]interface{})
	templateInput["JSON"] = template.JS(jsonData)
	templateInput["Legend"] = legend

	templateName := chartDisplay.TemplateName()
	err = tmpl.ExecuteTemplate(c.Response().Writer, templateName, templateInput)
	if err != nil {
		log.Printf("Error executing chart template: %v\n", err)
		return err
	}
	return nil
}
