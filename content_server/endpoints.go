package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type HandlerContext struct {
	EchoCtx echo.Context
	PGCtx   *PostgresContext
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

func (hCtx *HandlerContext) createAccount() error {
	user := getUser(hCtx.EchoCtx)
	passHash, msg, ok := hCtx.validateSignup(user)
	if !ok {
		return errorDiv(hCtx.EchoCtx, msg)
	}

	err := hCtx.PGCtx.insertAccount(user, passHash)
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

func (hCtx *HandlerContext) upload() error {
	// NOT IMPLEMENTED - endpoint for uploading files to storage.
	// Should use wrappers from filestore.go to abstract local vs S3 vs google etc
	// Should use JWT claims + postgres user to store file pointers and verify permissions
	// Also add in file parsers to extract raw text from

	fs := LocalStorage{
		Directory: "~/Documents/GoServer/filesystem",
	}

	// user := getUser(c)

	file, err := hCtx.EchoCtx.FormFile("file")
	if err != nil {
		return err
	}

	rawText, err := hCtx.EchoCtx.FormFile("raw_text")
	if err != nil {
		return err
	}


	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	Save(fs, rawText string, file io.Reader, filename string, pgContext *PostgresContext, uuid string)

	// // Use the storage interface to save the file
	// err = storage.Save(src, file.Filename)
	// if err != nil {
	// 	return err
	// }

	// // Additional logic (if any)

	// return nil
}
