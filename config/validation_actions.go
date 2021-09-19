package config

import (
	"errors"
	"fmt"

	"github.com/jxsl13/ocrmypdf-watchdog/internal"
	configo "github.com/jxsl13/simple-configo"
)

func dirMustExist(dir *string) configo.ActionFunc {
	return func() error {
		if dir == nil {
			return errors.New("nil pointer pased in dirMustEixst")
		}
		if !internal.IsExistingDir(*dir) {
			return fmt.Errorf("directory not found: %s", *dir)
		}
		return nil
	}
}
