package platform

import (
	"os"
	"path/filepath"
)

func GetExecPath() string {
	exec, err := os.Executable()
	if err != nil {
		panic(err)
	}

	if exec, err = filepath.EvalSymlinks(exec); err != nil {
		panic(err)
	}

	return exec
}

func GetExecDir() string {
	return filepath.Dir(GetExecPath())
}
