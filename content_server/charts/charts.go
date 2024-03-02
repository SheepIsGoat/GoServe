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
	ChartBuilder[query ChartQueryParams, raw ChartRawData, displayPtr ChartDisplay] struct {
		// Used to create a ChartDisplay object of the specified type, for rendering charts.js
		ChartProcessor ChartProcessor[query, raw, displayPtr]
		Tmpl           *template.Template
	}

	ChartProcessor[query ChartQueryParams, raw ChartRawData, displayPtr ChartDisplay] interface {
		FetchData(*pg.PostgresContext, query) ([]raw, error)
		PopulateDisplay([]raw) (displayPtr, error)
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

	chartDisplay, err := cb.ChartProcessor.PopulateDisplay(data)
	if err != nil {
		log.Printf("Failed to populate display for %T: %v", chartQuery, err)
		return err
	}
	// chartDisplay := cb.ChartProcessor._validateDisplayCast(_charDisp)

	jsonData, err := json.Marshal(chartDisplay)
	if err != nil {
		log.Printf("Failed to cast chart input as json for %T: %v", chartQuery, err)
		return err
	}

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
