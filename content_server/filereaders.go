package main

import (
	"fmt"
	"io"
	"strings"
)

func readTxt(file io.Reader) (string, error) {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, file)
	if err != nil {
		return "", fmt.Errorf("error reading .txt file: %w", err)
	}
	return buf.String(), nil
}
