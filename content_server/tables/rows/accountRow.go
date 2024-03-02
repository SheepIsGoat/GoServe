package rows

import (
	"fmt"
	"log"
	"time"

	"main/tables/cells"
	"main/tables/pagination"
	"main/templating/components"

	pg "main/postgres"
)

var _ RowProcessor[AccountRow] = AccountRowProcessor{}

type AccountRow struct {
	ProfileAvatar string
	ProfileName   string
	ProfileTitle  string
	Amount        float64
	Status        string
	Date          time.Time
}

func (ar AccountRow) _isRow() bool { return true }

type AccountRowProcessor struct{}

func (arp AccountRowProcessor) GetHeaders() []string {
	return []string{"Client", "Amount", "Status", "Date"}
}

func (arp AccountRowProcessor) BuildRowCells(ar AccountRow) []components.DivComponent {
	// populates cells with data based on populated AccountRow struct
	profile := components.DivComponent{
		Data: cells.ProfileCell{
			Avatar: ar.ProfileAvatar,
			Name:   ar.ProfileName,
			Title:  ar.ProfileTitle,
		},
	}
	amount := components.DivComponent{
		Data: cells.BasicCell{
			Val: fmt.Sprint(ar.Amount),
		},
	}
	_colorName := cells.StatusColorMap[ar.Status]
	_colorCss := cells.ColorCssMap[_colorName]
	status := components.DivComponent{
		Data: cells.StatusCell{
			Status: ar.Status,
			Color:  _colorCss,
		},
	}
	date := components.DivComponent{
		Data: cells.BasicCell{
			Val: ar.Date.Format("2006-01-02"),
		},
	}
	return []components.DivComponent{profile, amount, status, date}
}

func (arp AccountRowProcessor) QuerySQLToStructArray(pgContext *pg.PostgresContext, uuid string, pagination pagination.PaginConfig) ([]AccountRow, error) {
	// Do we need to generalize ORDER BY?
	query := `
	SELECT a.avatar, a.name, a.title, i.amount, i.status, i.date
	FROM "SampleAccounts" a
	JOIN "SampleInvoices" i ON a.id = i.account_id
	WHERE a.id IS NOT NULL
	ORDER BY i.date DESC
	LIMIT $1
	OFFSET $2
	`

	limit := pagination.ItemsPerPage
	offset := (pagination.CurrentPage - 1) * pagination.ItemsPerPage
	rows, err := pgContext.Pool.Query(pgContext.Ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	var results []AccountRow
	for rows.Next() {
		var ar AccountRow
		if err := rows.Scan(&ar.ProfileAvatar, &ar.ProfileName, &ar.ProfileTitle, &ar.Amount, &ar.Status, &ar.Date); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		results = append(results, ar)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

func (arp AccountRowProcessor) Count(pgContext *pg.PostgresContext, uuid string) (int, error) {
	query := `
	SELECT COUNT(*)
	FROM "SampleAccounts" a
	JOIN "SampleInvoices" i ON a.id = i.account_id
	`
	var count int
	rows, err := pgContext.Pool.Query(pgContext.Ctx, query)
	if err != nil {
		return count, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			return count, err
		}
	} else {
		return 0, fmt.Errorf("no rows returned")
	}
	return count, err
}
