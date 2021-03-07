package utils

import (
	"github.com/karrick/godirwalk"
	"os"
	"path/filepath"
	"strings"
)

type FileMap struct {
	path      string
	pathCache map[string]string
}

func NewFileMap(path string) *FileMap {
	fPath := strings.ReplaceAll(path, "\\", "/")
	if !strings.HasSuffix(fPath, "/") {
		fPath += "/"
	}

	fileMap := &FileMap{
		path: fPath,
		pathCache: make(map[string]string),
	}

	_ = godirwalk.Walk(fPath, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			fixedPath := strings.TrimPrefix(strings.ReplaceAll(osPathname, "\\", "/"), fPath)

			fileMap.pathCache[strings.ToLower(fixedPath)] = fixedPath

			return nil
		},
		Unsorted: true,
	})

	return fileMap
}

func (f *FileMap) GetFile(path string) (string, error) {
	sPath := strings.ToLower(f.path)
	fPath := strings.TrimPrefix(strings.ReplaceAll(strings.ToLower(path), "\\", "/"), sPath)

	if resolved, ok := f.pathCache[fPath]; ok {
		return filepath.Join(f.path, resolved), nil
	}

	return "", os.ErrNotExist
}