package config

import (
	"errors"
	"strconv"

	configo "github.com/jxsl13/simple-configo"
)

func OctalInt(out *uint32) configo.ParserFunc {
	return func(value string) error {
		if out == nil {
			return errors.New("out value must not be nil: OctalInt")
		}
		i, err := strconv.ParseUint(value, 8, 32)
		if err != nil {
			return err
		}
		*out = uint32(i)
		return nil
	}
}
