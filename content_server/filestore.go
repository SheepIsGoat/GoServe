package main

import "io"

//
// NEED to make sure users only have access to their files.
// we'll use database for that. But can use UID in filenames

type LocalStorage struct {
	Directory string
}

func (l *LocalStorage) Save(file io.Reader, filename string) error {
	// Logic to save file to local directory
}

// S3 storage
type S3Storage struct {
	BucketName string
	// Other AWS configurations
}

func (s *S3Storage) Save(file io.Reader, filename string) error {
	// Logic to save file to S3
}
