package files

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

func NewFileMap(path string) (*FileMap, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	fPath := strings.ReplaceAll(path, "\\", "/")
	if !strings.HasSuffix(fPath, "/") {
		fPath += "/"
	}

	fileMap := &FileMap{
		path:      fPath,
		pathCache: make(map[string]string),
	}

	_ = godirwalk.Walk(fPath, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			fixedPath := strings.TrimPrefix(strings.ReplaceAll(osPathname, "\\", "/"), fPath)

			fileMap.pathCache[strings.ToLower(fixedPath)] = fixedPath

			return nil
		},
		Unsorted:            true,
		FollowSymbolicLinks: true,
	})

	return fileMap, nil
}

func (f *FileMap) GetFile(path string) (string, error) {
	sPath := strings.ToLower(f.path)
	fPath := strings.TrimPrefix(strings.ReplaceAll(strings.ToLower(path), "\\", "/"), sPath)

	if resolved, ok := f.pathCache[fPath]; ok {
		return filepath.Join(f.path, resolved), nil
	}

	return "", os.ErrNotExist
}

func (f *FileMap) GetMap() map[string]string {
	retMap := make(map[string]string)

	for k, v := range f.pathCache {
		retMap[k] = filepath.Join(f.path, v)
	}

	return retMap
}
