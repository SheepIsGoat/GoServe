package rows

import (
	"fmt"
	"log"
	pg "main/postgres"
	"main/tables/cells"
	"main/tables/pagination"
	"main/templating/components"
	"path/filepath"
	"time"
)

type FileRow struct {
	Filename   string
	UploadTime time.Time
	FileExt    string
	RawText    string
	BucketDir  string
	Location   string
	FileURL    string
}

func (FileRow) _isRow() bool { return true }

type FileRowProcessor struct{}

func (frp FileRowProcessor) Count(pgContext *pg.PostgresContext, uuid string) (int, error) {
	query := `
	SELECT COUNT(*)
	FROM "files" f
	WHERE f.account_uuid = $1
	`
	var count int
	rows, err := pgContext.Pool.Query(pgContext.Ctx, query, uuid)
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

func (frp FileRowProcessor) QuerySQLToStructArray(pgContext *pg.PostgresContext, uuid string, pagination pagination.PaginConfig) ([]FileRow, error) {
	// Do we need to generalize ORDER BY?
	query := `
	SELECT f.unique_filename, f.upload_time, f.file_ext, f.raw_text, f.bucket_dir, f.location
	FROM "files" f
	WHERE f.account_uuid = $3
	LIMIT $1
	OFFSET $2
	`

	limit := pagination.ItemsPerPage
	offset := (pagination.CurrentPage - 1) * pagination.ItemsPerPage
	rows, err := pgContext.Pool.Query(pgContext.Ctx, query, limit, offset, uuid)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	var results []FileRow
	for rows.Next() {
		var fr FileRow
		if err := rows.Scan(&fr.Filename, &fr.UploadTime, &fr.FileExt, &fr.RawText, &fr.BucketDir, &fr.Location); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		fr.FileURL = filepath.Join(fr.BucketDir, fr.Filename)
		results = append(results, fr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

func (frp FileRowProcessor) BuildRowCells(fr FileRow) []components.DivComponent {
	filename := components.DivComponent{
		Data: cells.ModalCell{
			LinkText:     fr.Filename,
			ModalContent: fr.RawText,
		},
	}
	extension := components.DivComponent{
		Data: cells.BasicCell{
			Val: fr.FileExt,
		},
	}
	uploadTime := components.DivComponent{
		Data: cells.BasicCell{
			Val: fr.UploadTime.Format("2006-01-02 15:04:05"),
		},
	}
	return []components.DivComponent{filename, extension, uploadTime}
}

func (frp FileRowProcessor) GetHeaders() []string {
	return []string{"File", "Extension", "Upload Time"}
}
