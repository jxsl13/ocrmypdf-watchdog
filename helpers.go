package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/google/uuid"
)

func generateString(prefix string) string {
	uuid := uuid.New().String()

	return prefix + strings.Replace(uuid, "-", "", -1)[:32-len(prefix)]
}

func generateRandomUser() string {
	return generateString("u")
}

func generateRandomGroup() string {
	return generateString("g")
}

func fileInfo(filePath string) (string, error) {
	sb := strings.Builder{}

	info, err := os.Stat(filePath)

	if err != nil {
		return "", errFileNotFound
	}

	isDir := info.IsDir()
	if isDir {
		sb.WriteString("Permissions of folder: ")
	} else {
		sb.WriteString("Permissions of file: ")
	}
	sb.WriteString(fmt.Sprintf("%q\n", filePath))

	sb.WriteString("\tPerm: ")
	sb.WriteString(fmt.Sprint(info.Mode().Perm()))
	sb.WriteString("\n")

	sb.WriteString("\tSys: ")
	sb.WriteString(fmt.Sprint(info.Sys()))
	sb.WriteString("\n")

	var UID, GID int
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		UID = int(stat.Uid)
		GID = int(stat.Gid)
	} else {
		// we are not in linux, this won't work anyway in windows,
		// but maybe you want to log warnings
		UID = -1
		GID = -1
	}

	sb.WriteString(fmt.Sprintf("\tUID: %d\n\tGID: %d\n", UID, GID))
	return sb.String(), nil
}

func printInfo(filePath string) {

	info, err := fileInfo(filePath)
	if err != nil {
		log.Printf("Failed to open %q, %q", filePath, err)
		return
	}

	lines := strings.Split(info, "\n")

	for line := range lines {
		log.Println(line)
	}
}

func contains(element string, in []string) bool {
	for _, s := range in {
		if s == element {
			return true
		}
	}
	return false
}
