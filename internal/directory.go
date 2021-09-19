package internal

import "os"

// isExist checks is a directory exists
func IsExistingDir(dirPath string) bool {
	stat, err := os.Stat(dirPath)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return stat.IsDir()
}
