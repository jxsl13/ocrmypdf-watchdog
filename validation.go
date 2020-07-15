package main

import (
	"log"
	"net/http"
	"os"
)

// IsExist checks is a file or directory exists
func IsExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// IsPDF checks the content type of the file.
func IsPDF(filePath string) bool {
	contentType, err := GetFileContentType(filePath)
	if err != nil {
		log.Println("file type error:", err)
		return false
	}

	log.Printf("content type of %q is %q", filePath, contentType)
	if contentType != "application/pdf" {
		return false
	}
	return true
}

// GetFileContentType checks the file type.
func GetFileContentType(file string) (string, error) {

	// Open File
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err = f.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}
