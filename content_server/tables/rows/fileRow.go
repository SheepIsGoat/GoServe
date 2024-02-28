package rows

import "time"

// type RowProcessor[T Row] interface {
// 	Count(*pg.PostgresContext) (int, error)
// 	QuerySQLToStructArray(*pg.PostgresContext, pagination.PaginConfig) ([]T, error)
// 	BuildRowCells(T) []cells.TableCell
// 	GetHeaders() []string
// }

type FileRow struct {
	Filename   string
	UploadTime time.Time
	FileExt    string
	RawText    string
	BucketDir  string
	Location   string
	FileURL    string
}
