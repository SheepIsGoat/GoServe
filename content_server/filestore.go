package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	pg "main/postgres"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
)

// //

type Filesystem interface {
	Write(file io.Reader, filename string) error
	Delete(filename string) error
	GetStorageClass() *StorageClass
	GetLocation() string
}

type StorageClass struct {
	Config FileSystemConfig
}

type FileSystemConfig struct {
	BucketDir string
}

type FileInput struct {
	Filename    string
	AccountUUID string
	RawText     string
	File        io.Reader // optional, can be empty
}

type FileObject struct {
	FileId      string
	Filepath    string
	AccountUUID string
	UploadTime  time.Time
	Filename    string
	FileExt     string
	RawText     string
}

func _createFileInput(c echo.Context) (FileInput, error) {
	fileInput := FileInput{}

	file, err := c.FormFile("file")
	if err != nil {
		return fileInput, err
	}
	if file == nil {
		return fileInput, echo.NewHTTPError(http.StatusBadRequest, "No file in request")
	}
	fileInput.Filename = file.Filename

	src, err := file.Open()
	if err != nil {
		return fileInput, err
	}
	defer src.Close()

	fileInput.File = src

	fileInput.RawText = c.FormValue("raw_text")

	uuid, ok := c.Get("ID").(string)
	if !ok {
		return fileInput, echo.NewHTTPError(http.StatusUnauthorized, "User UUID not found in JWT claims")
	}
	fileInput.AccountUUID = uuid
	return fileInput, nil
}

// Saving files
func SaveFile(pgContext *pg.PostgresContext, filesystem Filesystem, fileInput FileInput) error {
	// Saves the file to disk on the filesystem of your choice, then indexes the result in postgres
	fileOutput, buf, err := fileInput.createFileOutput()
	if err != nil {
		log.Printf("Failed to create FileOutput: %v", err)
		return err
	}

	return fileOutput.writeFile(pgContext, filesystem, buf)
}

func (input FileInput) createFileOutput() (FileObject, bytes.Buffer, error) {
	fileOutput := FileObject{
		AccountUUID: input.AccountUUID,
	}

	fileOutput.Filename = input.Filename
	fileOutput.UploadTime = time.Now()
	fileOutput.Filepath = constructUniqueFilename(input.AccountUUID, fileOutput.UploadTime, input.Filename)

	var buf bytes.Buffer
	_, err := buf.ReadFrom(input.File)
	if err != nil && err != io.EOF {
		return fileOutput, buf, fmt.Errorf("error reading file into buffer: %w", err)
	}

	ext, err := getFileExt(input.Filename)
	if err != nil {
		return fileOutput, buf, err
	}
	fileOutput.FileExt = ext

	rawText := input.RawText
	if rawText == "" {
		if buf.Len() > 0 {
			rawText, err = extractText(buf, ext)
			if err != nil {
				return fileOutput, buf, err
			}
		} else {
			return fileOutput, buf, fmt.Errorf("empty file and no raw text? There's nothing here!")
		}
	}
	fileOutput.RawText = rawText
	return fileOutput, buf, nil
}

func (fo FileObject) writeFile(pgContext *pg.PostgresContext, filesystem Filesystem, buf bytes.Buffer) error {
	err := fo.Index(pgContext, filesystem)
	if err != nil {
		log.Printf("Error inserting file to table %v, %v", fo, err)
		return err
	}

	err = filesystem.Write(bytes.NewReader(buf.Bytes()), fo.Filepath)
	if err != nil {
		return fo.Delete(pgContext, filesystem, true)
	}

	return nil
}

func (fo FileObject) Delete(pgContext *pg.PostgresContext, filesystem Filesystem, force bool) error {
	if fo.Filepath == "" {
		sqlStatement := `SELECT filepath FROM files WHERE id = $1`

		rows, err := pgContext.Pool.Query(pgContext.Ctx, sqlStatement, fo.FileId)
		if err != nil {
			return fmt.Errorf("query execution error: %w", err)
		}
		defer rows.Close()

		var filepath string
		if rows.Next() {
			err = rows.Scan(&filepath)
			if err != nil {
				log.Printf("Failed to scan row: %v", err)
				return err
			}
		}
		fo.Filepath = filepath
	}
	var delErr error
	if fo.Filepath == "" {
		delErr = fmt.Errorf("No filepath specified, cannot delete file %s", fo.Filename)
	} else {
		delErr = filesystem.Delete(fo.Filepath)
	}
	if delErr != nil {
		log.Printf("Failed to delete file from disk: %v", delErr)
		if !force {
			return delErr
		}
	}
	dbErr := deleteFileRecord(pgContext, fo.FileId)
	if dbErr != nil {
		log.Printf("Failed to delete database record: %v", dbErr)
		return dbErr
	}
	if delErr != nil && force {
		return delErr
	}
	return nil
}

func deleteFileRecord(pgContext *pg.PostgresContext, fileId string) error {
	sqlStatement := `DELETE FROM files WHERE id = $1`
	_, err := pgContext.Pool.Exec(pgContext.Ctx, sqlStatement, fileId)
	if err != nil {
		return fmt.Errorf("error deleting file record: %w", err)
	}
	return nil
}

func constructUniqueFilename(accountUUID string, uploadTime time.Time, filename string) string {
	return fmt.Sprintf("%s-%v-%s", accountUUID, uploadTime, filename)
}

func getFileExt(filename string) (string, error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		return "", fmt.Errorf("file has no extension: %s", filename)
	}
	return ext, nil
}

func extractText(buf bytes.Buffer, ext string) (string, error) {
	switch ext {
	case ".txt":
		return readTxt(buf)
	case ".pdf":
		return readPdf(buf)
	case ".doc", ".docx":
		// Use a library or microservice to extract text from Word documents
	default:
		return "", fmt.Errorf("unsupported file extension")
	}
	return "Extracted text", nil
}

func (fo FileObject) Index(pgContext *pg.PostgresContext, filesystem Filesystem) error {
	storageClass := filesystem.GetStorageClass()
	sqlStatement := `
	INSERT INTO files (
		filename,
		filepath,
		account_uuid, 
		upload_time, 
		file_ext, 
		raw_text, 
		bucket_dir, 
		location
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := pgContext.Pool.Exec(
		pgContext.Ctx,
		sqlStatement,
		fo.Filename,
		fo.Filepath,
		fo.AccountUUID,
		fo.UploadTime,
		fo.FileExt,
		fo.RawText,
		storageClass.Config.BucketDir,
		filesystem.GetLocation(),
	)
	return err
}

type LocalStorage struct {
	StorageClass
}

var _ Filesystem = (*LocalStorage)(nil)

func (l *LocalStorage) Write(file io.Reader, filename string) error {
	fullPath := filepath.Join(l.StorageClass.Config.BucketDir, filename)
	outFile, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	return err
}

func (l *LocalStorage) Delete(filename string) error {
	fullPath := filepath.Join(l.StorageClass.Config.BucketDir, filename)
	return os.Remove(fullPath)
}

func (l *LocalStorage) GetStorageClass() *StorageClass {
	return &l.StorageClass
}

func (l *LocalStorage) GetLocation() string {
	return "local"
}

type S3Storage struct {
	StorageClass
}

func (s *S3Storage) Write(file io.Reader, filename string) error {
	// Logic to save file to S3
	return nil
}
