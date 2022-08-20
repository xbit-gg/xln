package util

import (
	"errors"
	"os"
)

func FileExists(file string) bool {
	_, err := os.Stat(file)
	return !errors.Is(err, os.ErrNotExist)
}
