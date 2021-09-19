package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// IsExist checks is a file or directory exists
func IsExistingFile(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

// IsPDF checks the content type of the file.
func IsPDF(filePath string) bool {
	// file name must end with .pdf
	if strings.ToLower(filepath.Ext(filePath)) != ".pdf" {
		return false
	}

	if !IsExistingFile(filePath) {
		return false
	}

	contentType, err := GetFileContentType(filePath)
	if err != nil {
		log.Println("file type error:", err)
		return false
	}

	return contentType == "application/pdf"
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
	b := [512]byte{}
	buffer := b[:]

	_, err = f.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}
