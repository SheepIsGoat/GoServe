package main

import (
	"fmt"
	"html/template"
	"log"
	"main/charts"
	pg "main/postgres"
	"main/tables"
	"main/tables/rows"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"
)

type HandlerContext struct {
	EchoCtx echo.Context
	PGCtx   *pg.PostgresContext
}

func errorDiv(c echo.Context, message string) error {
	errorMessageTemplate := `
    <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative" role="alert">
        <strong class="font-bold">Oops!</strong>
        <span class="block sm:inline">%s</span>
    </div>`
	errorMessage := fmt.Sprintf(errorMessageTemplate, message)
	return c.HTML(http.StatusOK, errorMessage)
}

func successDiv(c echo.Context, message string) error {
	errorMessageTemplate := `
    <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative" role="alert">
        <strong class="font-bold">Oops!</strong>
        <span class="block sm:inline">%s</span>
    </div>`
	errorMessage := fmt.Sprintf(errorMessageTemplate, message)
	return c.HTML(http.StatusOK, errorMessage)
}

func (hCtx *HandlerContext) createAccount() error {
	user := getUser(hCtx.EchoCtx)
	passHash, msg, ok := hCtx.validateSignup(user)
	if !ok {
		return errorDiv(hCtx.EchoCtx, msg)
	}

	err := insertAccount(hCtx.PGCtx, user, passHash)
	if err != nil {
		log.Print(err)
		return errorDiv(hCtx.EchoCtx, "Failed to create new account")
	}

	hCtx.EchoCtx.Response().Header().Set("HX-Redirect", "/app/")
	return nil
}

func (hCtx *HandlerContext) loginEndpoint() error {
	uid, errMsg, statusCode := hCtx.authenticateUser()
	statusOK := statusCode >= 200 && statusCode < 300
	if !statusOK {
		return errorDiv(hCtx.EchoCtx, errMsg)
	}

	ok := setCookie(hCtx.EchoCtx, uid)
	if !ok {
		return errorDiv(hCtx.EchoCtx, "Internal server error")
	}

	fmt.Printf("Successful login for user %v\n", uid)
	hCtx.EchoCtx.Response().Header().Set("HX-Redirect", "/app/")
	return nil
}

func (hCtx *HandlerContext) Upload() error {
	fileInput, err := _createFileInput(hCtx.EchoCtx)
	if err != nil {
		log.Printf("Failed to parse file upload: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Home directory not found")
		return err
	}

	filesystem := &LocalStorage{
		StorageClass{
			Config: FileSystemConfig{
				BucketDir: filepath.Join(homeDir, "/Documents/GoServer/filesystem"),
			},
		},
	}

	err = SaveFile(hCtx.PGCtx, filesystem, fileInput)
	if err != nil {
		log.Printf("Failed to save file; %v", err)
		return errorDiv(hCtx.EchoCtx, "Failed to upload file")
	}
	return successDiv(hCtx.EchoCtx, "Successfully uploaded file")
}

func (hCtx *HandlerContext) Table(tmpl *template.Template) error {
	// tableName := hCtx.EchoCtx.QueryParam("tableName")
	tableName := "Account Invoices"

	itemsPerPageStr := hCtx.EchoCtx.QueryParam("itemsPerPage")
	var err error
	itemsPerPage := uint64(10)
	if itemsPerPageStr != "" {
		itemsPerPage, err = strconv.ParseUint(itemsPerPageStr, 10, 32)
		if err != nil {
			log.Printf("Could not get itemsPerPage for table %s: %v\n", tableName, err)
			return err
		}
	}

	currentPageStr := hCtx.EchoCtx.QueryParam("page")
	currentPage := uint64(1)
	if currentPageStr != "" {
		currentPage, err = strconv.ParseUint(currentPageStr, 10, 32)
		if err != nil {
			log.Printf("Could not get currentPage for table %s: %v\n", tableName, err)
		}
	}

	processor := &rows.AccountRowProcessor{}
	totalRows, err := processor.Count(hCtx.PGCtx)
	if err != nil {
		log.Printf("Could not count rows for table %s: %v\n", tableName, err)
		return err
	}
	table := tables.Table[rows.AccountRow]{}
	table.Pagination.Init(
		tableName,
		uint32(totalRows),
		uint32(currentPage),
		uint32(itemsPerPage),
		7,
	)

	return table.RenderTable(hCtx.EchoCtx, hCtx.PGCtx, tmpl, processor)
}

func (hCtx *HandlerContext) PieChart(tmpl *template.Template) error {
	pieBuilder := charts.PieBuilder{
		ChartProcessor: charts.PieProcessor{},
		Tmpl:           tmpl,
	}

	pieQuery := charts.PieQuery{
		Table:   "SampleInvoices",
		Col:     "status",
		GroupBy: "status",
	}

	return pieBuilder.RenderChart(hCtx.EchoCtx, hCtx.PGCtx, tmpl, pieQuery)
}
