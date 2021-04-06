package file

import (
	"os"
	"path/filepath"
)

func AutoCreateFile(filePath string, perm os.FileMode) error {
	exist := func() bool {
		if _, err := os.Stat(filePath); err != nil {
			if os.IsExist(err) {
				return true
			}
			return false
		}
		return true
	}
	if !exist() {
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, perm); err != nil {
			return err
		}
		if _, err := os.Create(filePath); err != nil {
			return err
		}
		if err := os.Chmod(filePath, 0755); err != nil {
			return err
		}
	}
	return nil
}
