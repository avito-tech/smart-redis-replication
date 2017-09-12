package resp

import (
	"os"
)

// IsDir возвращает true если path это каталог
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
