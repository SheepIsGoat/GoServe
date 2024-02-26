package rows

import (
	"fmt"
	"log"
	"time"

	"main/tables/cells"
	"main/tables/pagination"

	pg "main/postgres"
)

type AccountRow struct {
	ProfileAvatar string
	ProfileName   string
	ProfileTitle  string
	Amount        float64
	Status        string
	Date          time.Time
}

type AccountRowProcessor struct{}

func (arp *AccountRowProcessor) GetHeaders() []string {
	return []string{"Client", "Amount", "Status", "Date"}
}

func (arp *AccountRowProcessor) BuildRowCells(ar AccountRow) []cells.TableCell {
	// populates cells with data based on populated AccountRow struct
	profile := cells.TableCell{
		Data: cells.ProfileCell{
			Avatar: ar.ProfileAvatar,
			Name:   ar.ProfileName,
			Title:  ar.ProfileTitle,
		},
	}
	amount := cells.TableCell{
		Data: cells.BasicCell{
			Val: fmt.Sprint(ar.Amount),
		},
	}
	_colorName := cells.StatusColorMap[ar.Status]
	_colorCss := cells.ColorCssMap[_colorName]
	status := cells.TableCell{
		Data: cells.StatusCell{
			Status: ar.Status,
			Color:  _colorCss,
		},
	}
	date := cells.TableCell{
		Data: cells.BasicCell{
			Val: ar.Date.Format("2006-01-02"),
		},
	}
	return []cells.TableCell{profile, amount, status, date}
}

func (arp *AccountRowProcessor) QuerySQLToStructArray(pgContext *pg.PostgresContext, pagination pagination.PaginConfig) ([]AccountRow, error) {
	// Do we need to generalize ORDER BY?
	query := `
	SELECT a.avatar, a.name, a.title, i.amount, i.status, i.date
	FROM "SampleAccounts" a
	JOIN "SampleInvoices" i ON a.id = i.account_id
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

func (arp *AccountRowProcessor) Count(pgContext *pg.PostgresContext) (int, error) {
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
