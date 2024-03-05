package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
)

func readTxt(buf bytes.Buffer) (string, error) {
	reader := bytes.NewReader(buf.Bytes())
	writer := new(strings.Builder)
	_, err := io.Copy(writer, reader)
	if err != nil {
		return "", fmt.Errorf("error reading .txt file: %w", err)
	}
	return buf.String(), nil
}

const extractPdfAPIEndpoint = "http://localhost:8000/extract/text"

func readPdf(buf bytes.Buffer) (string, error) {
	// Prepare a form that you will submit to your FastAPI server
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// part, err := writer.CreateFormFile("file", "filename.pdf")
	// if err != nil {
	// 	return "", err
	// }
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, "filename.pdf"))
	h.Set("Content-Type", "application/pdf")

	part, err := writer.CreatePart(h)
	if err != nil {
		return "", err
	}

	// Copy the PDF buffer to the multipart form
	_, err = part.Write(buf.Bytes())
	if err != nil {
		return "", err
	}
	writer.Close()

	// Create a HTTP client and post the request
	client := &http.Client{}
	req, err := http.NewRequest("POST", extractPdfAPIEndpoint, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Check if the response is successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to extract text, server responded with status code: %d", resp.StatusCode)
	}

	// Process the responseBody to extract the text
	// Assuming the response is a JSON with a key "extracted_text"
	var result map[string]string
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", err
	}
	extractedText, ok := result["extracted_text"]
	if !ok {
		return "", fmt.Errorf("failed to extract text, key 'extracted_text' not found in response")
	}

	return extractedText, nil
}
