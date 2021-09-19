package internal

import (
	"io/fs"
	"io/ioutil"
	"os"
)

func Copy(src, dest string, perms ...fs.FileMode) error {
	var perm fs.FileMode = 0644
	if len(perms) > 0 {
		perm = perms[0]
	}
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dest, input, perm)
}

func Move(src, dest string, perms ...fs.FileMode) error {
	err := Copy(src, dest, perms...)
	if err != nil {
		return err
	}
	return os.Remove(src)
}
