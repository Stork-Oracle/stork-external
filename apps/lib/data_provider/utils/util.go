package utils

import (
	"path/filepath"
	"runtime"
)

func GetCurrentDirName() string {
	_, file, _, ok := runtime.Caller(1) // 1 means the caller of this function
	if !ok {
		return ""
	}
	return filepath.Base(filepath.Dir(file))
}
