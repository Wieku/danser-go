package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type DirEntry fs.DirEntry

var SkipDir = errors.New("skip this directory")
var SkipChildDirs = errors.New("skip child directories")
var SkipAll = errors.New("skip everything and stop the walk")

type WalkDirFunc func(path string, level int, d DirEntry) error

func walkDir(path string, level int, d DirEntry, walkDirFn WalkDirFunc) error {
	skipChildDirs := false

	if err := walkDirFn(path, level, d); err != nil {
		if errors.Is(err, SkipChildDirs) {
			skipChildDirs = true
		} else {
			if errors.Is(err, SkipDir) {
				err = nil
			}

			return err
		}
	}

	dirs, files, _ := readDir(path)

	for _, d1 := range files {
		name1 := filepath.Join(path, d1.Name())
		if err := walkDirFn(name1, level, d1); err != nil {
			if errors.Is(err, SkipChildDirs) {
				skipChildDirs = true
			} else {
				return err
			}
		}
	}

	if skipChildDirs {
		return nil
	}

	for _, d1 := range dirs {
		name1 := filepath.Join(path, d1.Name())
		if err := walkDir(name1, level+1, d1, walkDirFn); err != nil {
			return err
		}
	}

	return nil
}

// WalkDir is similar to filepath.WalkDir, but it follows symlinks, passes tree depth to WalkDirFunc and allows to skip only subdirectories
func WalkDir(root string, fn WalkDirFunc) error {
	info, err := os.Stat(root)
	if err == nil {
		err = walkDir(root, 0, fs.FileInfoToDirEntry(info), fn)
	}
	if errors.Is(err, SkipDir) || errors.Is(err, SkipAll) {
		return nil
	}
	return err
}

// SearchFiles allows searching for files with wildcards, maxLevel sets the maximum tree depth, -1 is all depths, 0 is only in root etc
func SearchFiles(root string, pattern string, maxLevel int) (retString []string, err error) {
	err = WalkDir(root, func(path string, level int, de DirEntry) error {
		if maxLevel > -1 && level >= maxLevel && de.IsDir() {
			return SkipChildDirs
		}

		if !de.IsDir() {
			matchTest, err2 := filepath.Match(pattern, de.Name())

			if err2 == nil && matchTest {
				retString = append(retString, path)
			}
		}

		return nil
	})

	return
}

// os.ReadDir, but instead of sorting by name, it splits files and directories to separate slices
func readDir(path string) (dirs []DirEntry, files []DirEntry, err error) {
	f, err2 := os.Open(path)
	if err2 != nil {
		return nil, nil, err2
	}
	defer f.Close()

	res, err2 := f.ReadDir(-1)

	for _, dirEntry := range res {
		if dirEntry.Type()&fs.ModeSymlink > 0 {
			info, err := os.Stat(filepath.Join(path, dirEntry.Name()))
			if err != nil {
				continue
			}

			dirEntry = fs.FileInfoToDirEntry(info)
		}

		if dirEntry.IsDir() {
			dirs = append(dirs, dirEntry)
		} else {
			files = append(files, dirEntry)
		}
	}

	return dirs, files, err2
}
