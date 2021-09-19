package config

import (
	"errors"
	"fmt"
	"os"

	configo "github.com/jxsl13/simple-configo"
)

func dirMustExist(dir *string) configo.ActionFunc {
	return func() error {
		if dir == nil {
			return errors.New("nil pointer pased in dirMustEixst")
		}
		if !isExistingDir(*dir) {
			return fmt.Errorf("directory not found: %s", *dir)
		}
		return nil
	}
}

// isExist checks is a directory exists
func isExistingDir(dirPath string) bool {
	stat, err := os.Stat(dirPath)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return stat.IsDir()
}
