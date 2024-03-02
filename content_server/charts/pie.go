package charts

import (
	"fmt"
	"log"
	pg "main/postgres"
	"strings"
	"time"
)

type PieBuilder = ChartBuilder[PieQuery, PieRawData, *PieDisplay]

var _ ChartProcessor[PieQuery, PieRawData, *PieDisplay] = PieProcessor{}

type PieDisplay struct {
	Type    string         `json:"type"`
	Data    pieDisplayData `json:"data"`
	Options pieOptions     `json:"options"`
}

type pieDisplayData struct {
	Datasets []pieDatasets `json:"datasets"`
	Labels   []string      `json:"labels"`
}

type pieDatasets struct {
	Data            []int    `json:"data"`
	BackgroundColor []string `json:"backgroundColor"`
	Label           string   `json:"label"`
}

type pieOptions struct {
	Responsive       bool      `json:"responsive"`
	CutoutPercentage int       `json:"cutoutPercentage"`
	Legend           pieLegend `json:"legend"`
}

type pieLegend struct {
	Display bool `json:"display"`
}

func (pd *PieDisplay) GetLegend() (htmlLegend, error) {
	legend := htmlLegend{
		Title: "Invoices",
	}
	if len(pd.Data.Datasets) < 1 {
		return legend, fmt.Errorf("PieDisplay has no data. Cannot get legend")
	}
	for idx, label := range pd.Data.Labels {
		color := pd.Data.Datasets[0].BackgroundColor[idx]
		legend.Labels = append(
			legend.Labels,
			coloredItem{
				Label:    label,
				HexColor: color,
			},
		)
	}
	return legend, nil
}

var defaultPieOptions = pieOptions{
	Responsive:       true,
	CutoutPercentage: 80,
	Legend: pieLegend{
		Display: false,
	},
}

func (pd *PieDisplay) _isDisplay() bool { return true }

func (pd *PieDisplay) TemplateName() string {
	return "charts/Pie"
}

func (pd *PieDisplay) Init() {
	pd.Type = "doughnut"
	pd.Data.Datasets = []pieDatasets{
		{
			Label: "Dummy Label",
		},
	}
	pd.Options = defaultPieOptions
}

type PieQuery struct {
	Table   string
	Col     string
	GroupBy string
	Time    TimeQuery
}

type TimeQuery struct {
	OrderedCol string
	After      time.Time
	Before     time.Time
}

func (pq PieQuery) _isQueryParams() bool { return true }

type PieProcessor struct{}

func (pp PieProcessor) FetchData(pgContext *pg.PostgresContext, pq PieQuery) ([]PieRawData, error) {
	var args []interface{}

	where := ""
	if pq.Time.OrderedCol != "" {
		where = fmt.Sprintf("WHERE %s > $1 AND %s < $2", pq.Time.OrderedCol, pq.Time.OrderedCol)
		args = append(args, pq.Time.After, pq.Time.Before)
	}

	// cols := pq.Cols
	cols := []string{pq.Col}
	countCols := make([]string, len(cols))
	for i, col := range cols {
		countCols[i] = fmt.Sprintf("COUNT(t.\"%s\")", strings.TrimSpace(col))
	}

	Select := fmt.Sprintf(`SELECT %s, %s`, strings.Join(countCols, ", "), pq.GroupBy)
	from := fmt.Sprintf(`FROM "%s" t`, pq.Table)
	where = where
	groupby := fmt.Sprintf(`GROUP BY t."%s"`, pq.GroupBy)
	query := fmt.Sprintf(`%s %s %s %s`, Select, from, where, groupby)
	// SELECT COUNT(status), status FROM "SampleInvoices" GROUP BY $1
	// `
	var results []PieRawData
	rows, err := pgContext.Pool.Query(pgContext.Ctx, query, args...)
	if err != nil {
		log.Printf("query execution error: %v\n", err)
		return results, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pid PieRawData
		if err := rows.Scan(&pid.Data, &pid.Label); err != nil { // Adjust according to actual columns
			log.Printf("Failed to scan row: %v\n", err)
			continue
		}
		results = append(results, pid)
	}

	if err := rows.Err(); err != nil {
		return results, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// func (pp PieProcessor) PopulateDisplay(input []PieRawData) (PieDisplay, error) {
func (pp PieProcessor) PopulateDisplay(input []PieRawData) (*PieDisplay, error) {
	// This is a little awkward to do on single elements instead of iterating over a slice
	// But the awkwardness makes the typing system safer
	pd := new(PieDisplay)
	pd.Init()
	if len(pd.Data.Datasets) < 1 {
		return pd, fmt.Errorf("PieDisplay.Init() failed to initialize properly")
	}

	colorPalette := ColorPalettes["Dark-To-Light"]
	// colors := []string{"teal", "blue", "purple", "white", "green", "red", "orange"}
	for i, pid := range input {
		pd.Data.Datasets[0].Data = append(pd.Data.Datasets[0].Data, pid.Data)

		color := colorPalette[i%len(colorPalette)]
		pd.Data.Datasets[0].BackgroundColor = append(pd.Data.Datasets[0].BackgroundColor, color.HexCode)
		pd.Data.Labels = append(pd.Data.Labels, pid.Label)
	}
	return pd, nil
}

type PieRawData struct {
	Data  int
	Label string
}

func (pid PieRawData) _isChart() bool { return true }
