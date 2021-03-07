package utils

import (
	"github.com/karrick/godirwalk"
	"strings"
)

func GenerateFileMap(path string) (pathCache map[string]string) {
	fPath := strings.ReplaceAll(path, "\\", "/")
	if !strings.HasSuffix(fPath, "/") {
		fPath += "/"
	}

	_ = godirwalk.Walk(fPath, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			fixedPath := strings.TrimPrefix(strings.ReplaceAll(osPathname, "\\", "/"), fPath)

			pathCache[strings.ToLower(fixedPath)] = fixedPath

			return nil
		},
		Unsorted: true,
	})

	return
}
