package charts

import (
	"fmt"
	"log"
	pg "main/postgres"
	"time"
)

type PieBuilder = ChartBuilder[PieQuery, PieRawData, PieDisplay]

type PieDisplay struct {
	Type string `json:"type"`
	data struct {
		datasets []pieDatasets
		labels   []string
	}
	options pieOptions
}

type pieDatasets struct {
	data            []int
	backgroundColor []string
	label           string
}

type pieOptions struct {
	responsive       bool
	cutoutPercentage int
	legend           struct {
		display bool
	}
}

var defaultPieOptions = pieOptions{
	responsive:       true,
	cutoutPercentage: 80,
	legend: struct{ display bool }{
		display: false,
	},
}

func (pd *PieDisplay) _isDisplay() bool { return true }

func (pd *PieDisplay) TemplateName() string {
	return "charts/Pie"
}

func (pd *PieDisplay) Init() {
	pd.Type = "doughnut"
	pd.data.datasets = []pieDatasets{
		{
			label: "Dummy Label",
		},
	}
	pd.options = defaultPieOptions
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

func (pp PieProcessor) _processorType(PieRawData) PieRawData { return PieRawData{} }

func (pp PieProcessor) _validateQueryCast(pq PieQuery) ChartQueryParams {
	var iface ChartQueryParams = pq
	return iface
}

func (pp PieProcessor) _validateRawCast(prd PieRawData) ChartRawData {
	var iface ChartRawData = prd
	return iface
}

func (pp PieProcessor) _validateDisplayCast(pd *PieDisplay) ChartDisplay {
	var iface ChartDisplay = pd
	return iface
}

func (pp PieProcessor) FetchData(pgContext *pg.PostgresContext, pq PieQuery) ([]PieRawData, error) {
	// func (pp PieProcessor) FetchData(pgContext *pg.PostgresContext, cqp ChartQueryParams[PieRawData]) (Dummy[PieRawData], error) {
	// pq, ok := cqp.(PieQuery)
	// if !ok {
	// 	return fmt.Errorf("Failed to cast query as PieQuery")
	// }
	args := append([]interface{}{}, pq.GroupBy)

	// where := ""
	// if pq.Time.OrderedCol != "" {
	// 	where = fmt.Sprintf("WHERE %s > $2 AND %s < $3", pq.Time.OrderedCol, pq.Time.OrderedCol)
	// 	args = append(args, pq.Time.After, pq.Time.Before)
	// }

	// // cols := pq.Cols
	// cols := []string{pq.Col}
	// countCols := make([]string, len(cols))
	// for i, col := range cols {
	// 	countCols[i] = fmt.Sprintf("COUNT(t.\"%s\")", strings.TrimSpace(col))
	// }

	// query := fmt.Sprintf(`
	// SELECT %s
	// FROM "%s" t
	// %s
	// GROUP BY $1
	// `, strings.Join(countCols, ", "), pq.Table, where)
	query := `
	SELECT COUNT(status), status FROM "SampleInvoices" GROUP BY $1
	`

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

	log.Printf("Returning pie results %v\n", results)
	return results, nil
}

// func (pp PieProcessor) PopulateDisplay(input []PieRawData) (PieDisplay, error) {
func (pp PieProcessor) PopulateDisplay(input []PieRawData) (*PieDisplay, error) {
	// This is a little awkward to do on single elements instead of iterating over a slice
	// But the awkwardness makes the typing system safer
	pd := new(PieDisplay)
	pd.Init()
	if len(pd.data.datasets) < 1 {
		return pd, fmt.Errorf("PieDisplay.Init() failed to initialize properly")
	}

	colorPalette := ColorPalettes["Dark-To-Light"]
	for i, pid := range input {
		pd.data.datasets[0].data = append(pd.data.datasets[0].data, pid.Data)

		color := colorPalette[i%len(colorPalette)]
		pd.data.datasets[0].backgroundColor = append(pd.data.datasets[0].backgroundColor, color.HexCode)
		pd.data.labels = append(pd.data.labels, pid.Label)
	}
	return pd, nil
}

type PieRawData struct {
	Data   int
	Label  string
	IthVal int //for cycling colors
}

func (pid PieRawData) _isChart() bool { return true }
