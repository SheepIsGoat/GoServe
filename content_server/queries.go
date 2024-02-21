package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
)

// Content
type queryContext[T any, U any] struct {
	pgContext PostgresContext
	userId    string
	endpoint  string
	query     query[T, U] // T and U should be structs
}

type query[T any, U any] struct {
	params  T // should be a struct
	returns U // should be a struct
}

type ContentErrorCommonFields struct {
	Endpoint    string
	Message     string
	LowLevelErr error
}

////////////
// Errors //
////////////

// Error funcs
func PrintContentError(err ContentErrorCommonFields, showLow bool) string {
	message := fmt.Sprintf("Error retrieving content. Endpoint: %s. Message: %s", err.Endpoint, err.Message)
	if showLow && err.LowLevelErr != nil {
		message += err.LowLevelErr.Error()
	}
	return message
}

func mapErrorToStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch err.(type) {
	case *ContentBadArgsError:
		return http.StatusBadRequest
	case *ContentFailedQueryError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Error Types
type ContentBadArgsError struct {
	ContentErrorCommonFields
}

func (e ContentBadArgsError) Error() string {
	return PrintContentError(e.ContentErrorCommonFields, false)
}

type ContentFailedQueryError struct {
	ContentErrorCommonFields
	query string
}

func (e ContentFailedQueryError) Error() string {
	return PrintContentError(e.ContentErrorCommonFields, false)
}

//////////////////////
// Content Handlers //
//////////////////////

// newContent
type newContentT = queryContext[newContentParams, newContentRow]
type newContentQueryT = query[newContentParams, newContentRow]

type newContentParams struct {
	timeStr string
}

type newContentRow struct {
	Title      string `pgx:"title"`
	Summary    string `pgx:"summary"`
	Background string `pgx:"background"`
}

func (m *newContentRow) ScanRow(rows pgx.Rows) error {
	return rows.Scan(&m.Title, &m.Summary, &m.Background)
}

func newContentHandler(c echo.Context, pgContext PostgresContext) error {
	cat := newContentT{
		pgContext: pgContext,
		query: newContentQueryT{
			newContentParams{
				timeStr: c.Get("time").(string),
			},
			newContentRow{},
		},
	}
	data, err := newContent(&cat)
	if len(data) == 0 {
		return c.JSON(mapErrorToStatus(err), "No results found")
	}
	return c.JSON(mapErrorToStatus(err), data)
}

func newContent(qContext *newContentT) ([]newContentRow, error) {
	timeStr := qContext.query.params.timeStr

	timeObj, err := parseTimeNewContent(timeStr, qContext.endpoint)
	if err != nil {
		log.Println("Error parsing time")
		return []newContentRow{}, err
	}

	return queryNewContent(qContext, timeObj)
}

func parseTimeNewContent(timeStr string, endpoint string) (time.Time, error) {
	timeObj, err := time.Parse(timeFormat, timeStr)
	if err != nil {
		rError := ContentBadArgsError{
			ContentErrorCommonFields{
				Endpoint:    endpoint,
				Message:     "Could not parse time",
				LowLevelErr: err,
			},
		}
		return timeObj, rError
	}
	return timeObj, nil
}

func queryNewContent(qContext *newContentT, timeObj time.Time) ([]newContentRow, error) {
	pool := qContext.pgContext.pool
	ctx := qContext.pgContext.ctx
	queryStr := "SELECT title, summary, background_url FROM content WHERE datetime>$1"
	rows, err := pool.Query(ctx, queryStr, timeObj)
	if err != nil {
		log.Printf("Failed query. User: %v, Error: %v", qContext.userId, err.Error())
		rError := ContentFailedQueryError{
			ContentErrorCommonFields{
				Endpoint:    qContext.endpoint,
				Message:     "Query failed",
				LowLevelErr: err,
			},
			queryStr,
		}
		return []newContentRow{}, rError
	}
	defer rows.Close()

	var result []newContentRow
	for rows.Next() {
		var row newContentRow
		err := row.ScanRow(rows)
		if err != nil {
			rError := ContentFailedQueryError{
				ContentErrorCommonFields{
					Endpoint:    qContext.endpoint,
					Message:     "Failed to read in query results",
					LowLevelErr: err,
				},
				queryStr,
			}
			return result, rError
		}
		result = append(result, row)
	}

	log.Printf("Query retrieved %v results", len(result))

	return result, nil
}

// other content handlers...
