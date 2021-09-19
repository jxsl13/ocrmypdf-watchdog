package internal

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func FileInfo(filePath string) (string, error) {
	sb := strings.Builder{}

	info, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}

	isDir := info.IsDir()
	if isDir {
		sb.WriteString("Permissions of folder: ")
	} else {
		sb.WriteString("Permissions of file: ")
	}
	sb.WriteString(fmt.Sprintf("%q\n", filePath))

	sb.WriteString("\tPerm: ")
	sb.WriteString(strconv.FormatUint(uint64(info.Mode().Perm()), 8))
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

	sb.WriteString(fmt.Sprintf("\tUID: %d\n\tGID: %d", UID, GID))
	return sb.String(), nil
}

func PrintInfo(filePath string) {

	info, err := FileInfo(filePath)
	if err != nil {
		log.Printf("Failed to open %q, %q", filePath, err)
		return
	}

	lines := strings.Split(info, "\n")
	for _, line := range lines {
		log.Println(line)
	}
}
