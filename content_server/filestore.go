package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// //
type FileSystemConfig struct {
	BucketDir string
	Location  string
}

type Filesystem interface {
	Write(file io.Reader, filename string) error
	Delete(filename string) error
	GetStorageClass() *StorageClass
}

type StorageClass struct {
	Config FileSystemConfig
}

type FileInput struct {
	Filename    string
	AccountUUID string
	RawText     string
	File        io.Reader // optional, can be empty
}

type FileOutput struct {
	AccountUUID    string
	UploadTime     int64
	UniqueFilename string
	FileExt        string
	RawText        string
}

// Saving files
func Save(pgContext *PostgresContext, filesystem Filesystem, fileInput FileInput) error {
	fileOutput, buf, err := fileInput.createFileOutput()
	if err != nil {
		return err
	}

	return fileOutput.writeFile(pgContext, filesystem, buf)
}

func (input FileInput) createFileOutput() (FileOutput, bytes.Buffer, error) {
	fileOutput := FileOutput{
		AccountUUID: input.AccountUUID,
	}

	fileOutput.UploadTime = time.Now().UnixNano()

	uniqueFilename := input.getUniqueFilename(fileOutput.UploadTime)
	fileOutput.UniqueFilename = uniqueFilename

	var buf bytes.Buffer
	_, err := buf.ReadFrom(input.File)
	if err != nil && err != io.EOF {
		return fileOutput, buf, fmt.Errorf("error reading file into buffer: %w", err)
	}

	ext, err := getFileExt(fileOutput.FileExt)
	if err != nil {
		return fileOutput, buf, err
	}
	fileOutput.FileExt = ext

	rawText := input.RawText
	if rawText == "" {
		if buf.Len() > 0 {
			rawText, err = extractText(bytes.NewReader(buf.Bytes()), ext)
			if err != nil {
				return fileOutput, buf, err
			}
		} else {
			return fileOutput, buf, fmt.Errorf("Empty file and no raw text? There's nothing here!")
		}
	}
	fileOutput.RawText = rawText
	return fileOutput, buf, nil
}

func (fo FileOutput) writeFile(pgContext *PostgresContext, filesystem Filesystem, buf bytes.Buffer) error {
	err := indexFile(pgContext, filesystem, fo)
	if err != nil {
		return err
	}

	err = filesystem.Write(bytes.NewReader(buf.Bytes()), fo.UniqueFilename)
	if err != nil {
		// Not atomic since it uses a filesystem and a database!
		delErr := filesystem.Delete(fo.UniqueFilename)
		if delErr != nil {
			log.Printf("Failed to delete file after write error: %v", delErr)
		}
		dbErr := deleteFileRecord(pgContext, fo.UniqueFilename)
		if dbErr != nil {
			log.Printf("Failed to delete database record after write error: %v", dbErr)
		}
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (input FileInput) getUniqueFilename(uploadTime int64) string {
	return fmt.Sprintf("%s-%v-%s", input.AccountUUID, uploadTime, input.Filename)
}

func getFileExt(filename string) (string, error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		return "", fmt.Errorf("file has no extension")
	}
	return ext, nil
}

func extractText(file io.Reader, ext string) (string, error) {
	// Pseudocode, as implementation depends on file type and libraries used
	switch ext {
	case ".txt":
		return readTxt(file)
	case ".pdf":
		// Use a PDF library to extract text
	case ".doc", ".docx":
		// Use a library to extract text from Word documents
	default:
		return "", fmt.Errorf("unsupported file extension")
	}
	return "Extracted text", nil
}

func indexFile(pgContext *PostgresContext, filesystem Filesystem, details FileOutput) error {
	storageClass := filesystem.GetStorageClass()
	sqlStatement := `
	INSERT INTO files (
		unique_filename, 
		account_uuid, 
		upload_time, 
		file_ext, 
		raw_text, 
		bucket_dir, 
		location
	) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := pgContext.pool.Exec(
		pgContext.ctx,
		sqlStatement,
		details.UniqueFilename,
		details.AccountUUID,
		details.UploadTime,
		details.FileExt,
		details.RawText,
		storageClass.Config.BucketDir,
		storageClass.Config.Location,
	)
	return err
}

func deleteFileRecord(pgContext *PostgresContext, uniqueFilename string) error {
	sqlStatement := `DELETE FROM files WHERE storage_filename = $1`
	_, err := pgContext.pool.Exec(pgContext.ctx, sqlStatement, uniqueFilename)
	if err != nil {
		return fmt.Errorf("error deleting file record: %w", err)
	}
	return nil
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

type S3Storage struct {
	StorageClass
}

func (s *S3Storage) Write(file io.Reader, filename string) error {
	// Logic to save file to S3
	return nil
}
