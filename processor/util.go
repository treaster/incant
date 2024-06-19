package processor

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

func FindFiles(fileRoot string) []string {
	var files []string
	filepath.WalkDir(fileRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println(err.Error())
		}
		baseName := filepath.Base(path)
		if baseName[0] == '.' {
			fmt.Printf("skipping dotfile %s\n", path)
			return nil
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	return files
}
